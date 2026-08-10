package sqlite

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

const insightCreateRetryLimit = 12

// CreateInsight は同一requestを冪等に知見として保存する。
func (s *Store) CreateInsight(ctx context.Context, requestID string, evidences []domain.InsightEvidence, statement, conditions, gaps string) (domain.Insight, bool, error) {
	fingerprint := insightFingerprint(evidences, statement, conditions, gaps)
	for attempt := 0; attempt < insightCreateRetryLimit; attempt++ {
		insight, created, createErr := s.createInsightOnce(ctx, requestID, evidences, statement, conditions, gaps, fingerprint)
		if createErr == nil || !isSQLiteBusy(createErr) && !isInsightOperationUnique(createErr) {
			return insight, created, createErr
		}
		replayed, found, replayErr := s.findInsightOperation(ctx, requestID)
		if replayErr == nil && found {
			if replayed.fingerprint != fingerprint {
				return domain.Insight{}, false, apperr.New(apperr.CodeInsightCreateRequestConflict)
			}
			return replayed.insight, false, nil
		}
		if replayErr != nil && !isSQLiteBusy(replayErr) {
			return domain.Insight{}, false, replayErr
		}
		if err := waitDraftSaveRetry(ctx, time.Duration(attempt+1)*time.Millisecond); err != nil {
			return domain.Insight{}, false, err
		}
	}

	return domain.Insight{}, false, fmt.Errorf("create insight retry exhausted")
}

// createInsightOnce は一つのtransactionで知見を保存する。
func (s *Store) createInsightOnce(ctx context.Context, requestID string, evidences []domain.InsightEvidence, statement, conditions, gaps, fingerprint string) (domain.Insight, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Insight{}, false, err
	}
	defer func() { _ = tx.Rollback() }()

	replayed, found, err := findInsightOperation(ctx, tx, requestID)
	if err != nil {
		return domain.Insight{}, false, err
	}
	if found {
		if replayed.fingerprint != fingerprint {
			return domain.Insight{}, false, apperr.New(apperr.CodeInsightCreateRequestConflict)
		}
		return replayed.insight, false, nil
	}
	if len(evidences) < 2 {
		return domain.Insight{}, false, apperr.New(apperr.CodeInsightCreateEvidenceInsufficient)
	}
	for _, evidence := range evidences {
		var exists bool
		if err := tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM experiment_conclusions WHERE id=? AND experiment_id=? AND state='finalized')", evidence.ConclusionID, evidence.ExperimentID).Scan(&exists); err != nil {
			return domain.Insight{}, false, err
		}
		if !exists {
			return domain.Insight{}, false, apperr.New(apperr.CodeInsightCreateEvidenceNotFound)
		}
	}

	id, err := newBriefingIdentifier()
	if err != nil {
		return domain.Insight{}, false, err
	}
	createdAt := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, "INSERT INTO insights (id,statement,applicability_conditions,verification_gaps,created_at) VALUES(?,?,?,?,?)", id, statement, conditions, gaps, createdAt.Format(time.RFC3339Nano)); err != nil {
		return domain.Insight{}, false, err
	}
	for sequence, evidence := range evidences {
		if _, err := tx.ExecContext(ctx, "INSERT INTO insight_evidences (insight_id,experiment_id,conclusion_id,sequence_no) VALUES(?,?,?,?)", id, evidence.ExperimentID, evidence.ConclusionID, sequence+1); err != nil {
			return domain.Insight{}, false, err
		}
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO insight_create_operations(request_id,request_fingerprint,insight_id) VALUES(?,?,?)", requestID, fingerprint, id); err != nil {
		return domain.Insight{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Insight{}, false, err
	}

	return domain.Insight{RequestID: requestID, InsightID: id, Evidences: evidences, Statement: statement, ApplicabilityConditions: conditions, VerificationGaps: gaps, CreatedAt: createdAt}, true, nil
}

type insightOperation struct {
	fingerprint string
	insight     domain.Insight
}
type insightOperationQuerier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

// findInsightOperation はrequestの保存済み知見を読み出す。
func (s *Store) findInsightOperation(ctx context.Context, requestID string) (insightOperation, bool, error) {
	return findInsightOperation(ctx, s.db, requestID)
}

// findInsightOperation はtransactionまたはDBから冪等結果を読み出す。
func findInsightOperation(ctx context.Context, q insightOperationQuerier, requestID string) (insightOperation, bool, error) {
	var operation insightOperation
	var createdAt string
	err := q.QueryRowContext(ctx, "SELECT o.request_fingerprint,o.insight_id,i.statement,i.applicability_conditions,i.verification_gaps,i.created_at FROM insight_create_operations o JOIN insights i ON i.id=o.insight_id WHERE o.request_id=?", requestID).Scan(&operation.fingerprint, &operation.insight.InsightID, &operation.insight.Statement, &operation.insight.ApplicabilityConditions, &operation.insight.VerificationGaps, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return insightOperation{}, false, nil
	}
	if err != nil {
		return insightOperation{}, false, err
	}
	operation.insight.RequestID = requestID
	operation.insight.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return insightOperation{}, false, err
	}
	rows, err := q.QueryContext(ctx, "SELECT experiment_id,conclusion_id FROM insight_evidences WHERE insight_id=? ORDER BY sequence_no", operation.insight.InsightID)
	if err != nil {
		return insightOperation{}, false, err
	}
	defer func() { _ = rows.Close() }()
	operation.insight.Evidences = make([]domain.InsightEvidence, 0)
	for rows.Next() {
		var evidence domain.InsightEvidence
		if err := rows.Scan(&evidence.ExperimentID, &evidence.ConclusionID); err != nil {
			return insightOperation{}, false, err
		}
		operation.insight.Evidences = append(operation.insight.Evidences, evidence)
	}
	if err := rows.Err(); err != nil {
		return insightOperation{}, false, err
	}
	return operation, true, nil
}

// insightFingerprint は正規payloadの安定した指紋を返す。
func insightFingerprint(evidences []domain.InsightEvidence, statement, conditions, gaps string) string {
	payload := struct {
		Evidences  []domain.InsightEvidence `json:"evidences"`
		Statement  string                   `json:"statement"`
		Conditions string                   `json:"conditions"`
		Gaps       string                   `json:"gaps"`
	}{evidences, statement, conditions, gaps}
	encoded, _ := json.Marshal(payload)
	hash := sha256.Sum256(encoded)
	return hex.EncodeToString(hash[:])
}

// isInsightOperationUnique は冪等operationの一意制約競合を判定する。
func isInsightOperationUnique(err error) bool {
	return strings.Contains(err.Error(), "insight_create_operations.request_id")
}
