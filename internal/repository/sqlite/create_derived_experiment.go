package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

const derivedExperimentRetryLimit = 14

var marshalDerivedExperimentPayload = json.Marshal

// CreateDerivedExperiment は適格な派生元からpreparing実験を原子的に作る。
func (s *Store) CreateDerivedExperiment(ctx context.Context, requestID, sourceID string, changes domain.DerivedExperimentChanges, reason string) (domain.DerivedExperiment, bool, error) {
	return s.createDerivedExperimentWithRetry(ctx, requestID, sourceID, changes, reason)
}

// createDerivedExperimentWithRetry はSQLiteの一時競合を待機して再試行する。
func (s *Store) createDerivedExperimentWithRetry(ctx context.Context, requestID, sourceID string, changes domain.DerivedExperimentChanges, reason string) (domain.DerivedExperiment, bool, error) {
	payload, err := canonicalDerivedExperimentPayload(changes, reason)
	if err != nil {
		return domain.DerivedExperiment{}, false, err
	}

	var lastErr error
	for attempt := range derivedExperimentRetryLimit {
		existing, found, findErr := s.findDerivedExperiment(ctx, requestID)
		if isSQLiteBusy(findErr) {
			lastErr = findErr
		} else if findErr != nil {
			return domain.DerivedExperiment{}, false, findErr
		} else if found {
			return derivedExperimentReplay(existing, sourceID, payload)
		} else {
			derived, created, createErr := s.createDerivedExperiment(ctx, requestID, sourceID, changes, payload)
			if !isSQLiteBusy(createErr) {
				return derived, created, createErr
			}
			lastErr = createErr
		}
		if attempt == derivedExperimentRetryLimit-1 {
			break
		}
		if err := waitDraftSaveRetry(ctx, time.Millisecond<<attempt); err != nil {
			return domain.DerivedExperiment{}, false, err
		}
	}

	return domain.DerivedExperiment{}, false, lastErr
}

// createDerivedExperiment は一回のtransactionで派生実験とその冪等snapshotを保存する。
func (s *Store) createDerivedExperiment(ctx context.Context, requestID, sourceID string, changes domain.DerivedExperimentChanges, payload string) (domain.DerivedExperiment, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.DerivedExperiment{}, false, fmt.Errorf("begin derived experiment: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	existing, found, err := findDerivedExperimentOperation(ctx, tx, requestID)
	if err != nil {
		return domain.DerivedExperiment{}, false, err
	}
	if found {
		return derivedExperimentReplay(existing, sourceID, payload)
	}

	source, err := findDerivedExperimentSource(ctx, tx, sourceID)
	if err != nil {
		return domain.DerivedExperiment{}, false, err
	}
	if !hasMaterialDerivedExperimentChange(source, changes) {
		return domain.DerivedExperiment{}, false, apperr.New(apperr.CodeDerivedExperimentInvalid)
	}

	id, err := newBriefingIdentifier()
	if err != nil {
		return domain.DerivedExperiment{}, false, fmt.Errorf("generate derived experiment ID: %w", err)
	}
	now := time.Now().UTC()
	nowValue := now.Format(time.RFC3339Nano)
	if _, err = tx.ExecContext(ctx, "INSERT INTO experiments (id,purpose,state,derived_from_experiment_id,updated_at) VALUES (?,?,?, ?,?)", id, source.Purpose, "preparing", sourceID, nowValue); err != nil {
		return domain.DerivedExperiment{}, false, fmt.Errorf("insert derived experiment: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO experiment_preparations
		(experiment_id, briefing_session_id, briefing_version_id, hypothesis, environment_conditions, initial_input, evaluation_criteria, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, id, source.BriefingSessionID, source.BriefingVersionID, source.Hypothesis, source.EnvironmentConditions, source.InitialInput, source.EvaluationAxes, nowValue, nowValue); err != nil {
		return domain.DerivedExperiment{}, false, fmt.Errorf("copy derived preparation: %w", err)
	}
	if err = applyDerivedExperimentChanges(ctx, tx, id, changes); err != nil {
		return domain.DerivedExperiment{}, false, err
	}
	if changes.Prompts == nil {
		changes.Prompts = &source.Prompts
	}
	for _, prompt := range *changes.Prompts {
		if _, err = tx.ExecContext(ctx, "INSERT INTO experiment_preparation_prompts (experiment_id, sequence_no, content) VALUES (?, ?, ?)", id, prompt.SequenceNo, prompt.Content); err != nil {
			return domain.DerivedExperiment{}, false, fmt.Errorf("copy derived prompt: %w", err)
		}
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO experiment_derived_operations
		(request_id,source_experiment_id,experiment_id,canonical_payload,created_at) VALUES (?,?,?,?,?)`, requestID, sourceID, id, payload, nowValue); err != nil {
		if isDerivedExperimentRequestConflict(err) {
			if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
				return domain.DerivedExperiment{}, false, fmt.Errorf("rollback derived experiment request conflict: %w", rollbackErr)
			}
			existing, found, findErr := s.findDerivedExperiment(ctx, requestID)
			if findErr != nil {
				return domain.DerivedExperiment{}, false, findErr
			}
			if found {
				return derivedExperimentReplay(existing, sourceID, payload)
			}
		}
		return domain.DerivedExperiment{}, false, fmt.Errorf("insert derived experiment operation: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return domain.DerivedExperiment{}, false, fmt.Errorf("commit derived experiment: %w", err)
	}

	return domain.DerivedExperiment{RequestID: requestID, ExperimentID: id, SourceExperimentID: sourceID, State: "preparing", CreatedAt: now}, true, nil
}

type derivedExperimentSource struct {
	Purpose               string
	Hypothesis            *string
	EnvironmentConditions string
	InitialInput          string
	EvaluationAxes        string
	Prompts               []domain.ExperimentPreparationPrompt
	BriefingSessionID     string
	BriefingVersionID     string
}

func findDerivedExperimentSource(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, sourceID string) (derivedExperimentSource, error) {
	var source derivedExperimentSource
	var fixedID, conclusionID string
	var hypothesis sql.NullString
	err := queryer.QueryRowContext(ctx, `SELECT COALESCE(c.id, ''), COALESCE(x.id, ''), COALESCE(c.purpose, ''), c.hypothesis, COALESCE(c.environment_conditions, ''), COALESCE(c.initial_input, ''), COALESCE(c.evaluation_axes, ''), COALESCE(p.briefing_session_id, ''), COALESCE(p.briefing_version_id, '')
		FROM experiments e
		LEFT JOIN experiment_fixed_conditions c ON c.id=e.fixed_condition_id
		LEFT JOIN experiment_conclusions x ON x.experiment_id=e.id AND x.state='finalized'
		LEFT JOIN experiment_preparations p ON p.experiment_id=e.id
		WHERE e.id=?`, sourceID).Scan(&fixedID, &conclusionID, &source.Purpose, &hypothesis, &source.EnvironmentConditions, &source.InitialInput, &source.EvaluationAxes, &source.BriefingSessionID, &source.BriefingVersionID)
	if err == sql.ErrNoRows {
		return derivedExperimentSource{}, apperr.New(apperr.CodeDerivedExperimentSourceNotFound)
	}
	if err != nil {
		return derivedExperimentSource{}, fmt.Errorf("find derived source: %w", err)
	}
	if fixedID == "" || conclusionID == "" {
		return derivedExperimentSource{}, apperr.New(apperr.CodeDerivedExperimentSourceNotEligible)
	}
	if hypothesis.Valid {
		source.Hypothesis = &hypothesis.String
	}
	rows, err := queryer.(interface {
		QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	}).QueryContext(ctx, "SELECT sequence_no, content FROM experiment_fixed_condition_prompts WHERE fixed_condition_id=? ORDER BY sequence_no", fixedID)
	if err != nil {
		return derivedExperimentSource{}, fmt.Errorf("find derived source prompts: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var prompt domain.ExperimentPreparationPrompt
		if err = rows.Scan(&prompt.SequenceNo, &prompt.Content); err != nil {
			return derivedExperimentSource{}, fmt.Errorf("scan derived source prompt: %w", err)
		}
		source.Prompts = append(source.Prompts, prompt)
	}
	if err = rows.Err(); err != nil {
		return derivedExperimentSource{}, fmt.Errorf("iterate derived source prompts: %w", err)
	}
	return source, nil
}

func applyDerivedExperimentChanges(ctx context.Context, tx *sql.Tx, experimentID string, changes domain.DerivedExperimentChanges) error {
	for _, change := range []struct {
		query string
		value *string
	}{
		{"UPDATE experiments SET purpose=? WHERE id=?", changes.Purpose},
		{"UPDATE experiment_preparations SET hypothesis=? WHERE experiment_id=?", changes.Hypothesis},
		{"UPDATE experiment_preparations SET environment_conditions=? WHERE experiment_id=?", changes.EnvironmentConditions},
		{"UPDATE experiment_preparations SET initial_input=? WHERE experiment_id=?", changes.InitialInput},
		{"UPDATE experiment_preparations SET evaluation_criteria=? WHERE experiment_id=?", changes.EvaluationAxes},
	} {
		if change.value == nil {
			continue
		}
		if _, err := tx.ExecContext(ctx, change.query, *change.value, experimentID); err != nil {
			return fmt.Errorf("apply derived experiment change: %w", err)
		}
	}
	return nil
}

func hasMaterialDerivedExperimentChange(source derivedExperimentSource, changes domain.DerivedExperimentChanges) bool {
	return (changes.Purpose != nil && *changes.Purpose != source.Purpose) ||
		(changes.Hypothesis != nil && !sameOptionalString(changes.Hypothesis, source.Hypothesis)) ||
		(changes.EnvironmentConditions != nil && *changes.EnvironmentConditions != source.EnvironmentConditions) ||
		(changes.InitialInput != nil && *changes.InitialInput != source.InitialInput) ||
		(changes.EvaluationAxes != nil && *changes.EvaluationAxes != source.EvaluationAxes) ||
		(changes.Prompts != nil && !slices.EqualFunc(*changes.Prompts, source.Prompts, func(a, b domain.ExperimentPreparationPrompt) bool {
			return a.SequenceNo == b.SequenceNo && a.Content == b.Content
		}))
}

type derivedExperimentPayload struct {
	Changes domain.DerivedExperimentChanges `json:"changes"`
	Reason  string                          `json:"reason"`
}

func canonicalDerivedExperimentPayload(changes domain.DerivedExperimentChanges, reason string) (string, error) {
	encoded, err := marshalDerivedExperimentPayload(derivedExperimentPayload{Changes: changes, Reason: reason})
	if err != nil {
		return "", fmt.Errorf("marshal derived experiment payload: %w", err)
	}
	return string(encoded), nil
}

func derivedExperimentReplay(existing derivedExperimentOperation, sourceID, payload string) (domain.DerivedExperiment, bool, error) {
	if existing.SourceExperimentID != sourceID || existing.CanonicalPayload != payload {
		return domain.DerivedExperiment{}, false, apperr.New(apperr.CodeDerivedExperimentRequestConflict)
	}
	return existing.DerivedExperiment, false, nil
}

type derivedExperimentOperation struct {
	domain.DerivedExperiment
	CanonicalPayload string
}

func (s *Store) findDerivedExperiment(ctx context.Context, requestID string) (derivedExperimentOperation, bool, error) {
	return findDerivedExperimentOperation(ctx, s.db, requestID)
}

func findDerivedExperimentOperation(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, requestID string) (derivedExperimentOperation, bool, error) {
	var operation derivedExperimentOperation
	var createdAt string
	err := queryer.QueryRowContext(ctx, "SELECT request_id, experiment_id, source_experiment_id, canonical_payload, created_at FROM experiment_derived_operations WHERE request_id=?", requestID).Scan(&operation.RequestID, &operation.ExperimentID, &operation.SourceExperimentID, &operation.CanonicalPayload, &createdAt)
	if err == sql.ErrNoRows {
		return derivedExperimentOperation{}, false, nil
	}
	if err != nil {
		return derivedExperimentOperation{}, false, fmt.Errorf("find derived experiment operation: %w", err)
	}
	if operation.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt); err != nil {
		return derivedExperimentOperation{}, false, fmt.Errorf("parse derived experiment creation time: %w", err)
	}
	operation.State = "preparing"
	return operation, true, nil
}

func isDerivedExperimentRequestConflict(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed: experiment_derived_operations.request_id")
}
