package sqlite

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

var readBriefingRandom = rand.Read

// briefingTransaction は実験ブリーフ記録のトランザクション境界。
type briefingTransaction interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) briefingRow
	Commit() error
	Rollback() error
}

// briefingRow はトランザクション内の単一行読み出し境界。
type briefingRow interface {
	Scan(...any) error
}

// sqliteBriefingTransaction はdatabase/sqlトランザクションのadapter。
type sqliteBriefingTransaction struct {
	tx *sql.Tx
}

// ExecContext はSQL実行を委譲。
func (t sqliteBriefingTransaction) ExecContext(ctx context.Context, query string, arguments ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, arguments...)
}

// QueryRowContext は単一行読み出しを委譲。
func (t sqliteBriefingTransaction) QueryRowContext(ctx context.Context, query string, arguments ...any) briefingRow {
	return t.tx.QueryRowContext(ctx, query, arguments...)
}

// Commit はトランザクション確定を委譲。
func (t sqliteBriefingTransaction) Commit() error {
	return t.tx.Commit()
}

// Rollback はトランザクション取消を委譲。
func (t sqliteBriefingTransaction) Rollback() error {
	return t.tx.Rollback()
}

// BeginExperimentBriefing は開始意図と準備セッションを原子的に保存。
func (s *Store) BeginExperimentBriefing(ctx context.Context, requestID string) (domain.ExperimentBriefingStart, bool, error) {
	existing, found, err := s.findExperimentBriefing(ctx, requestID)
	if err != nil {
		return domain.ExperimentBriefingStart{}, false, err
	}
	if found {
		return existing, false, nil
	}

	sessionID, err := newBriefingIdentifier()
	if err != nil {
		return domain.ExperimentBriefingStart{}, false, err
	}
	operationID, err := newBriefingIdentifier()
	if err != nil {
		return domain.ExperimentBriefingStart{}, false, err
	}
	start := domain.ExperimentBriefingStart{
		RequestID:         requestID,
		BriefingSessionID: sessionID,
		OperationID:       operationID,
		State:             domain.BriefingStartStateStarting,
	}

	tx, err := s.beginBriefingTransaction(ctx)
	if err != nil {
		return domain.ExperimentBriefingStart{}, false, fmt.Errorf("begin experiment briefing: %w", err)
	}
	createdAt := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, "INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", start.BriefingSessionID, "experiment_brief", start.State, createdAt, createdAt); err != nil {
		_ = tx.Rollback()

		return domain.ExperimentBriefingStart{}, false, fmt.Errorf("insert preparation session: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO briefing_operations (id, request_id, preparation_session_id, state) VALUES (?, ?, ?, ?)", start.OperationID, start.RequestID, start.BriefingSessionID, start.State); err != nil {
		_ = tx.Rollback()
		if isBriefingRequestConflict(err) {
			existing, found, findErr := s.findExperimentBriefing(ctx, requestID)
			if findErr != nil {
				return domain.ExperimentBriefingStart{}, false, findErr
			}
			if found {
				return existing, false, nil
			}
		}

		return domain.ExperimentBriefingStart{}, false, fmt.Errorf("insert briefing operation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.ExperimentBriefingStart{}, false, fmt.Errorf("commit experiment briefing: %w", err)
	}

	return start, true, nil
}

// BeginExperimentBriefMessage は開始済みsessionに紐付く送信意図を原子的に保存。
func (s *Store) BeginExperimentBriefMessage(ctx context.Context, requestID, briefingSessionID string) (domain.ExperimentBriefingMessageOperation, bool, error) {
	existing, found, err := s.findExperimentBriefMessageOperation(ctx, requestID)
	if err != nil {
		return domain.ExperimentBriefingMessageOperation{}, false, err
	}
	if found {
		if existing.BriefingSessionID != briefingSessionID {
			return domain.ExperimentBriefingMessageOperation{}, false, fmt.Errorf("briefing message request belongs to another session")
		}

		return existing, false, nil
	}

	operationID, err := newBriefingIdentifier()
	if err != nil {
		return domain.ExperimentBriefingMessageOperation{}, false, err
	}
	operation := domain.ExperimentBriefingMessageOperation{
		RequestID:         requestID,
		BriefingSessionID: briefingSessionID,
		OperationID:       operationID,
		State:             domain.BriefingStartStateStarting,
	}

	tx, err := s.beginBriefingTransaction(ctx)
	if err != nil {
		return domain.ExperimentBriefingMessageOperation{}, false, fmt.Errorf("begin briefing message: %w", err)
	}
	var state string
	if err := tx.QueryRowContext(ctx, "SELECT state FROM preparation_sessions WHERE id = ? AND kind = ?", briefingSessionID, "experiment_brief").Scan(&state); err != nil {
		_ = tx.Rollback()
		if err == sql.ErrNoRows {
			return domain.ExperimentBriefingMessageOperation{}, false, apperr.New(apperr.CodeBriefingNotFound)
		}

		return domain.ExperimentBriefingMessageOperation{}, false, fmt.Errorf("find briefing message session: %w", err)
	}
	if state != domain.BriefingStartStateStarted {
		_ = tx.Rollback()

		return domain.ExperimentBriefingMessageOperation{}, false, apperr.New(apperr.CodeBriefingNotActive)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO briefing_message_operations (id, request_id, preparation_session_id, state) VALUES (?, ?, ?, ?)", operation.OperationID, operation.RequestID, operation.BriefingSessionID, operation.State); err != nil {
		_ = tx.Rollback()
		if isBriefingMessageRequestConflict(err) {
			existing, found, findErr := s.findExperimentBriefMessageOperation(ctx, requestID)
			if findErr != nil {
				return domain.ExperimentBriefingMessageOperation{}, false, findErr
			}
			if found && existing.BriefingSessionID == briefingSessionID {
				return existing, false, nil
			}
		}

		return domain.ExperimentBriefingMessageOperation{}, false, fmt.Errorf("insert briefing message operation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.ExperimentBriefingMessageOperation{}, false, fmt.Errorf("commit briefing message: %w", err)
	}

	return operation, true, nil
}

// CompleteExperimentBriefMessage は安全な会話と下書き候補を送信完了として保存。
func (s *Store) CompleteExperimentBriefMessage(ctx context.Context, requestID, message string, result domain.ExperimentBriefingMessageResult) error {
	s.briefingMessageMu.Lock()
	defer s.briefingMessageMu.Unlock()

	operation, found, err := s.findExperimentBriefMessageOperation(ctx, requestID)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("complete briefing message: request not found")
	}

	tx, err := s.beginBriefingTransaction(ctx)
	if err != nil {
		return fmt.Errorf("begin briefing message completion: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	var nextSequence int
	if err := tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(sequence_no), 0) + 1 FROM briefing_messages WHERE preparation_session_id = ?", operation.BriefingSessionID).Scan(&nextSequence); err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("find next briefing message sequence: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO briefing_messages (preparation_session_id, sequence_no, role, content, created_at) VALUES (?, ?, ?, ?, ?)", operation.BriefingSessionID, nextSequence, "user", message, now); err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("insert user briefing message: %w", err)
	}
	if result.AssistantMessage != "" {
		if _, err := tx.ExecContext(ctx, "INSERT INTO briefing_messages (preparation_session_id, sequence_no, role, content, created_at) VALUES (?, ?, ?, ?, ?)", operation.BriefingSessionID, nextSequence+1, "assistant", result.AssistantMessage, now); err != nil {
			_ = tx.Rollback()

			return fmt.Errorf("insert assistant briefing message: %w", err)
		}
	}
	if result.Brief != nil {
		if err := s.insertExperimentBriefVersion(ctx, tx, operation.BriefingSessionID, *result.Brief, now); err != nil {
			_ = tx.Rollback()

			return err
		}
	}
	if _, err := tx.ExecContext(ctx, "UPDATE briefing_message_operations SET state = ?, failure_code = '' WHERE request_id = ?", domain.BriefingStartStateStarted, requestID); err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("complete briefing message operation: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "UPDATE preparation_sessions SET updated_at = ? WHERE id = ?", now, operation.BriefingSessionID); err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("update briefing message session: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit briefing message completion: %w", err)
	}

	return nil
}

// FailExperimentBriefMessage は安全な送信失敗を保存。
func (s *Store) FailExperimentBriefMessage(ctx context.Context, requestID, failureCode string) error {
	result, err := s.failBriefingMessageOperation(ctx, requestID, failureCode)
	if err != nil {
		return fmt.Errorf("fail briefing message operation: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count failed briefing message operations: %w", err)
	}
	if affected != 1 {
		return fmt.Errorf("fail briefing message operation: request not found")
	}

	return nil
}

// findExperimentBriefMessageOperation はrequest IDに対応する送信結果を取得。
func (s *Store) findExperimentBriefMessageOperation(ctx context.Context, requestID string) (domain.ExperimentBriefingMessageOperation, bool, error) {
	operation := domain.ExperimentBriefingMessageOperation{RequestID: requestID}
	err := s.db.QueryRowContext(ctx, "SELECT preparation_session_id, id, state, failure_code FROM briefing_message_operations WHERE request_id = ?", requestID).Scan(&operation.BriefingSessionID, &operation.OperationID, &operation.State, &operation.FailureCode)
	if err == sql.ErrNoRows {
		return domain.ExperimentBriefingMessageOperation{}, false, nil
	}
	if err != nil {
		return domain.ExperimentBriefingMessageOperation{}, false, fmt.Errorf("find briefing message operation: %w", err)
	}

	return operation, true, nil
}

// insertExperimentBriefVersion は構造化済みのブリーフ候補を一版として保存。
func (s *Store) insertExperimentBriefVersion(ctx context.Context, tx briefingTransaction, briefingSessionID string, brief domain.ExperimentBrief, createdAt string) error {
	versionID, err := newBriefingIdentifier()
	if err != nil {
		return fmt.Errorf("generate briefing version identifier: %w", err)
	}
	var nextVersion int
	if err := tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(version_no), 0) + 1 FROM briefing_versions WHERE preparation_session_id = ?", briefingSessionID).Scan(&nextVersion); err != nil {
		return fmt.Errorf("find next briefing version: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO briefing_versions (id, preparation_session_id, version_no, decision, hypothesis, success_criteria, required_conditions, open_question, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", versionID, briefingSessionID, nextVersion, brief.Decision, brief.Hypothesis, brief.SuccessCriteria, brief.RequiredConditions, brief.OpenQuestion, createdAt); err != nil {
		return fmt.Errorf("insert briefing version: %w", err)
	}

	return nil
}

// GetExperimentBriefing は保存済み実験ブリーフを会話順で読み出す。
func (s *Store) GetExperimentBriefing(ctx context.Context, briefingSessionID string) (domain.ExperimentBriefing, bool, error) {
	briefing, found, err := s.findExperimentBriefingSession(ctx, briefingSessionID)
	if err != nil || !found {
		return domain.ExperimentBriefing{}, found, err
	}

	messages, latestMessageAt, err := s.listExperimentBriefingMessages(ctx, briefingSessionID)
	if err != nil {
		return domain.ExperimentBriefing{}, false, err
	}
	latestBrief, latestBriefAt, err := s.findLatestExperimentBrief(ctx, briefingSessionID)
	if err != nil {
		return domain.ExperimentBriefing{}, false, err
	}
	briefing.Messages = messages
	briefing.LatestBrief = latestBrief
	briefing.LastConfirmedAt = latestTime(briefing.LastConfirmedAt, latestMessageAt, latestBriefAt)

	return briefing, true, nil
}

// findExperimentBriefingSession は開始済み実験ブリーフセッションを取得。
func (s *Store) findExperimentBriefingSession(ctx context.Context, briefingSessionID string) (domain.ExperimentBriefing, bool, error) {
	var briefing domain.ExperimentBriefing
	var createdAt string
	var updatedAt string
	err := s.db.QueryRowContext(ctx, "SELECT state, created_at, updated_at FROM preparation_sessions WHERE id = ? AND kind = ?", briefingSessionID, "experiment_brief").Scan(&briefing.State, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return domain.ExperimentBriefing{}, false, nil
	}
	if err != nil {
		return domain.ExperimentBriefing{}, false, fmt.Errorf("find briefing session: %w", err)
	}
	if updatedAt == "" {
		updatedAt = createdAt
	}
	confirmedAt, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return domain.ExperimentBriefing{}, false, fmt.Errorf("parse briefing session update time: %w", err)
	}
	briefing.LastConfirmedAt = confirmedAt.UTC()

	return briefing, true, nil
}

// listExperimentBriefingMessages は利用者表示用会話を時系列順に取得。
func (s *Store) listExperimentBriefingMessages(ctx context.Context, briefingSessionID string) (messages []domain.ExperimentBriefingMessage, latestAt time.Time, err error) {
	rows, err := s.db.QueryContext(ctx, "SELECT role, content, sequence_no, created_at FROM briefing_messages WHERE preparation_session_id = ? ORDER BY sequence_no ASC", briefingSessionID)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("query briefing messages: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close briefing message rows: %w", closeErr)
		}
	}()

	messages = make([]domain.ExperimentBriefingMessage, 0)
	for rows.Next() {
		var message domain.ExperimentBriefingMessage
		var createdAt string
		if err := rows.Scan(&message.Role, &message.Content, &message.SequenceNo, &createdAt); err != nil {
			return nil, time.Time{}, fmt.Errorf("scan briefing message: %w", err)
		}
		message.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, time.Time{}, fmt.Errorf("parse briefing message creation time: %w", err)
		}
		message.CreatedAt = message.CreatedAt.UTC()
		latestAt = latestTime(latestAt, message.CreatedAt)
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, time.Time{}, fmt.Errorf("iterate briefing messages: %w", err)
	}

	return messages, latestAt, nil
}

// findLatestExperimentBrief は最新の保存済みブリーフ版を取得。
func (s *Store) findLatestExperimentBrief(ctx context.Context, briefingSessionID string) (*domain.ExperimentBrief, time.Time, error) {
	brief := &domain.ExperimentBrief{}
	var hypothesis sql.NullString
	var openQuestion sql.NullString
	var createdAt string
	err := s.db.QueryRowContext(ctx, "SELECT id, decision, hypothesis, success_criteria, required_conditions, open_question, created_at FROM briefing_versions WHERE preparation_session_id = ? ORDER BY version_no DESC LIMIT 1", briefingSessionID).Scan(&brief.VersionID, &brief.Decision, &hypothesis, &brief.SuccessCriteria, &brief.RequiredConditions, &openQuestion, &createdAt)
	if err == sql.ErrNoRows {
		return nil, time.Time{}, nil
	}
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("find latest briefing version: %w", err)
	}
	if hypothesis.Valid {
		brief.Hypothesis = &hypothesis.String
	}
	if openQuestion.Valid {
		brief.OpenQuestion = &openQuestion.String
	}
	parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("parse briefing version creation time: %w", err)
	}

	return brief, parsedCreatedAt.UTC(), nil
}

// latestTime はゼロ値を除いて最も新しいUTC時刻を返す。
func latestTime(times ...time.Time) time.Time {
	latest := time.Time{}
	for _, candidate := range times {
		if candidate.After(latest) {
			latest = candidate.UTC()
		}
	}

	return latest
}

// MarkExperimentBriefingStarted は開始済み状態を保存。
func (s *Store) MarkExperimentBriefingStarted(ctx context.Context, requestID string) error {
	return s.updateExperimentBriefing(ctx, requestID, domain.BriefingStartStateStarted, "")
}

// MarkExperimentBriefingFailed は安全な失敗コードを保存。
func (s *Store) MarkExperimentBriefingFailed(ctx context.Context, requestID, failureCode string) error {
	return s.updateExperimentBriefing(ctx, requestID, domain.BriefingStartStateFailed, failureCode)
}

// findExperimentBriefing はrequest IDに対応する開始結果を取得。
func (s *Store) findExperimentBriefing(ctx context.Context, requestID string) (domain.ExperimentBriefingStart, bool, error) {
	start := domain.ExperimentBriefingStart{RequestID: requestID}
	err := s.db.QueryRowContext(ctx, "SELECT preparation_session_id, id, state, failure_code FROM briefing_operations WHERE request_id = ?", requestID).Scan(&start.BriefingSessionID, &start.OperationID, &start.State, &start.FailureCode)
	if err == sql.ErrNoRows {
		return domain.ExperimentBriefingStart{}, false, nil
	}
	if err != nil {
		return domain.ExperimentBriefingStart{}, false, fmt.Errorf("find briefing operation: %w", err)
	}

	return start, true, nil
}

// updateExperimentBriefing は開始状態を関連レコードへ同期。
func (s *Store) updateExperimentBriefing(ctx context.Context, requestID, state, failureCode string) error {
	tx, err := s.beginBriefingTransaction(ctx)
	if err != nil {
		return fmt.Errorf("begin briefing update: %w", err)
	}
	result, err := tx.ExecContext(ctx, "UPDATE briefing_operations SET state = ?, failure_code = ? WHERE request_id = ?", state, failureCode, requestID)
	if err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("update briefing operation: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("count briefing operation updates: %w", err)
	}
	if affected != 1 {
		_ = tx.Rollback()

		return fmt.Errorf("update briefing operation: request not found")
	}
	updatedAt := time.Now().UTC().Format(time.RFC3339Nano)
	sessionResult, err := tx.ExecContext(ctx, "UPDATE preparation_sessions SET state = ?, updated_at = ? WHERE id = (SELECT preparation_session_id FROM briefing_operations WHERE request_id = ?)", state, updatedAt, requestID)
	if err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("update preparation session: %w", err)
	}
	sessionAffected, err := sessionResult.RowsAffected()
	if err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("count preparation session updates: %w", err)
	}
	if sessionAffected != 1 {
		_ = tx.Rollback()

		return fmt.Errorf("update preparation session: session not found")
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit briefing update: %w", err)
	}

	return nil
}

// isBriefingRequestConflict はrequest IDの一意制約競合を判定。
func isBriefingRequestConflict(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed: briefing_operations.request_id")
}

// isBriefingMessageRequestConflict は送信request IDの一意制約競合を判定。
func isBriefingMessageRequestConflict(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed: briefing_message_operations.request_id")
}

// newBriefingIdentifier は外部識別子を生成。
func newBriefingIdentifier() (string, error) {
	bytes := make([]byte, 16)
	if _, err := readBriefingRandom(bytes); err != nil {
		return "", fmt.Errorf("read random identifier: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

// ListExperiments は取消済みを分離して実験一覧を読み出す。
func (s *Store) ListExperiments(ctx context.Context) (domain.ExperimentCollection, error) {
	experiments, err := s.listByCancellation(ctx, false)
	if err != nil {
		return domain.ExperimentCollection{}, err
	}

	cancelledExperiments, err := s.listByCancellation(ctx, true)
	if err != nil {
		return domain.ExperimentCollection{}, err
	}

	lastConfirmedAt, err := s.confirmList(ctx)
	if err != nil {
		return domain.ExperimentCollection{}, err
	}

	return domain.ExperimentCollection{
		Experiments:          experiments,
		CancelledExperiments: cancelledExperiments,
		LastConfirmedAt:      lastConfirmedAt,
	}, nil
}

// listByCancellation は取消状態で実験を絞り込む。
func (s *Store) listByCancellation(ctx context.Context, cancelled bool) (experiments []domain.Experiment, err error) {
	operator := "<>"
	if cancelled {
		operator = "="
	}

	query := "SELECT id, purpose, state, progress_summary, derived_from_experiment_id, updated_at FROM experiments WHERE state " + operator + " ? ORDER BY updated_at DESC, id ASC"
	rows, err := s.db.QueryContext(ctx, query, "cancelled")
	if err != nil {
		return nil, fmt.Errorf("query experiments: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close experiments rows: %w", closeErr)
		}
	}()

	experiments = make([]domain.Experiment, 0)
	for rows.Next() {
		experiment, err := scanExperiment(rows)
		if err != nil {
			return nil, err
		}

		experiments = append(experiments, experiment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate experiments: %w", err)
	}

	return experiments, nil
}

// scanExperiment はSQLite行をdomain実験へ変換。
func scanExperiment(rows *sql.Rows) (domain.Experiment, error) {
	var experiment domain.Experiment
	var derivedFromExperimentID sql.NullString
	var updatedAt string
	if err := rows.Scan(
		&experiment.ID,
		&experiment.Purpose,
		&experiment.State,
		&experiment.ProgressSummary,
		&derivedFromExperimentID,
		&updatedAt,
	); err != nil {
		return domain.Experiment{}, fmt.Errorf("scan experiment: %w", err)
	}

	parsedUpdatedAt, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return domain.Experiment{}, fmt.Errorf("parse experiment update time: %w", err)
	}
	experiment.UpdatedAt = parsedUpdatedAt
	if derivedFromExperimentID.Valid {
		experiment.DerivedFromExperimentID = &derivedFromExperimentID.String
	}

	return experiment, nil
}

// confirmList は成功した一覧取得時刻を記録して返す。
func (s *Store) confirmList(ctx context.Context) (*time.Time, error) {
	confirmedAt := time.Now().UTC()
	if _, err := s.db.ExecContext(ctx, "INSERT INTO application_metadata (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value", "last_confirmed_at", confirmedAt.Format(time.RFC3339Nano)); err != nil {
		return nil, fmt.Errorf("record list confirmation: %w", err)
	}

	return &confirmedAt, nil
}
