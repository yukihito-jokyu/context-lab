package sqlite

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
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

// BeginStopExperimentBriefing は停止意図を原子的に保存する。
func (s *Store) BeginStopExperimentBriefing(ctx context.Context, requestID, briefingSessionID string) (domain.ExperimentBriefingStopOperation, bool, error) {
	existing, found, err := s.findStopExperimentBriefing(ctx, requestID)
	if err != nil {
		return domain.ExperimentBriefingStopOperation{}, false, err
	}
	if found {
		if existing.BriefingSessionID != briefingSessionID {
			return domain.ExperimentBriefingStopOperation{}, false, apperr.New(apperr.CodeBriefingRequestInvalid)
		}

		return existing, false, nil
	}

	operationID, err := newBriefingIdentifier()
	if err != nil {
		return domain.ExperimentBriefingStopOperation{}, false, err
	}
	operation := domain.ExperimentBriefingStopOperation{
		RequestID:         requestID,
		BriefingSessionID: briefingSessionID,
		OperationID:       operationID,
		State:             domain.BriefingStartStateStarting,
	}

	tx, err := s.beginBriefingTransaction(ctx)
	if err != nil {
		return domain.ExperimentBriefingStopOperation{}, false, fmt.Errorf("begin stop experiment briefing: %w", err)
	}
	var state string
	if err := tx.QueryRowContext(ctx, "SELECT state FROM preparation_sessions WHERE id = ? AND kind = ?", briefingSessionID, "experiment_brief").Scan(&state); err != nil {
		_ = tx.Rollback()
		if err == sql.ErrNoRows {
			return domain.ExperimentBriefingStopOperation{}, false, apperr.New(apperr.CodeBriefingNotFound)
		}

		return domain.ExperimentBriefingStopOperation{}, false, fmt.Errorf("find stop briefing session: %w", err)
	}
	if state != domain.BriefingStartStateStarted {
		_ = tx.Rollback()

		return domain.ExperimentBriefingStopOperation{}, false, apperr.New(apperr.CodeBriefingNotActive)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO briefing_stop_operations (id, request_id, preparation_session_id, state) VALUES (?, ?, ?, ?)", operation.OperationID, operation.RequestID, operation.BriefingSessionID, operation.State); err != nil {
		_ = tx.Rollback()
		if isBriefingStopRequestConflict(err) {
			existing, found, findErr := s.findStopExperimentBriefing(ctx, requestID)
			if findErr != nil {
				return domain.ExperimentBriefingStopOperation{}, false, findErr
			}
			if found && existing.BriefingSessionID == briefingSessionID {
				return existing, false, nil
			}
		}

		return domain.ExperimentBriefingStopOperation{}, false, fmt.Errorf("insert briefing stop operation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.ExperimentBriefingStopOperation{}, false, fmt.Errorf("commit stop experiment briefing: %w", err)
	}

	return operation, true, nil
}

// CompleteStopExperimentBriefing は停止確認後にsessionを停止済みとして保存する。
func (s *Store) CompleteStopExperimentBriefing(ctx context.Context, requestID string) error {
	return s.updateStopExperimentBriefing(ctx, requestID, domain.BriefingStartStateStopped, "")
}

// FailStopExperimentBriefing は安全な停止失敗を保存する。
func (s *Store) FailStopExperimentBriefing(ctx context.Context, requestID, failureCode string) error {
	return s.updateStopExperimentBriefing(ctx, requestID, domain.BriefingStartStateFailed, failureCode)
}

// findStopExperimentBriefing はrequest IDに対応する停止結果を取得する。
func (s *Store) findStopExperimentBriefing(ctx context.Context, requestID string) (domain.ExperimentBriefingStopOperation, bool, error) {
	operation := domain.ExperimentBriefingStopOperation{RequestID: requestID}
	err := s.db.QueryRowContext(ctx, "SELECT preparation_session_id, id, state, failure_code FROM briefing_stop_operations WHERE request_id = ?", requestID).Scan(&operation.BriefingSessionID, &operation.OperationID, &operation.State, &operation.FailureCode)
	if err == sql.ErrNoRows {
		return domain.ExperimentBriefingStopOperation{}, false, nil
	}
	if err != nil {
		return domain.ExperimentBriefingStopOperation{}, false, fmt.Errorf("find briefing stop operation: %w", err)
	}

	return operation, true, nil
}

// updateStopExperimentBriefing は停止operationとsession状態を同期する。
func (s *Store) updateStopExperimentBriefing(ctx context.Context, requestID, state, failureCode string) error {
	tx, err := s.beginBriefingTransaction(ctx)
	if err != nil {
		return fmt.Errorf("begin briefing stop update: %w", err)
	}
	result, err := tx.ExecContext(ctx, "UPDATE briefing_stop_operations SET state = ?, failure_code = ? WHERE request_id = ?", state, failureCode, requestID)
	if err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("update briefing stop operation: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("count briefing stop operation updates: %w", err)
	}
	if affected != 1 {
		_ = tx.Rollback()

		return fmt.Errorf("update briefing stop operation: request not found")
	}
	if state == domain.BriefingStartStateStopped {
		updatedAt := time.Now().UTC().Format(time.RFC3339Nano)
		sessionResult, err := tx.ExecContext(ctx, "UPDATE preparation_sessions SET state = ?, updated_at = ? WHERE id = (SELECT preparation_session_id FROM briefing_stop_operations WHERE request_id = ?)", state, updatedAt, requestID)
		if err != nil {
			_ = tx.Rollback()

			return fmt.Errorf("update stopped briefing session: %w", err)
		}
		sessionAffected, err := sessionResult.RowsAffected()
		if err != nil {
			_ = tx.Rollback()

			return fmt.Errorf("count stopped briefing session updates: %w", err)
		}
		if sessionAffected != 1 {
			_ = tx.Rollback()

			return fmt.Errorf("update stopped briefing session: session not found")
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit briefing stop update: %w", err)
	}

	return nil
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
	candidatePrompts, err := json.Marshal(brief.CandidatePrompts)
	if err != nil {
		return fmt.Errorf("marshal candidate prompts: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO briefing_versions (id, preparation_session_id, version_no, decision, hypothesis, success_criteria, required_conditions, open_question, created_at, purpose, candidate_prompts, evaluation_criteria, environment_conditions, initial_input) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", versionID, briefingSessionID, nextVersion, brief.Decision, brief.Hypothesis, brief.SuccessCriteria, brief.RequiredConditions, brief.OpenQuestion, createdAt, brief.Purpose, string(candidatePrompts), brief.EvaluationCriteria, brief.EnvironmentConditions, brief.InitialInput); err != nil {
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

// GetExperimentPreparation は準備中実験の編集条件を読み出す。
func (s *Store) GetExperimentPreparation(ctx context.Context, experimentID string) (preparation domain.ExperimentPreparation, found bool, err error) {
	preparation = domain.ExperimentPreparation{ExperimentID: experimentID}
	var hypothesis sql.NullString
	var updatedAt string
	err = s.db.QueryRowContext(ctx, "SELECT e.state, e.purpose, p.hypothesis, p.environment_conditions, p.initial_input, p.evaluation_criteria, p.briefing_version_id, b.decision, p.updated_at FROM experiments e JOIN experiment_preparations p ON p.experiment_id = e.id JOIN briefing_versions b ON b.id = p.briefing_version_id WHERE e.id = ?", experimentID).Scan(&preparation.State, &preparation.Purpose, &hypothesis, &preparation.EnvironmentConditions, &preparation.InitialInput, &preparation.EvaluationAxes, &preparation.Source.VersionID, &preparation.Source.State, &updatedAt)
	if err == sql.ErrNoRows {
		return domain.ExperimentPreparation{}, false, nil
	}
	if err != nil {
		return domain.ExperimentPreparation{}, false, fmt.Errorf("find experiment preparation: %w", err)
	}
	if hypothesis.Valid {
		preparation.Hypothesis = &hypothesis.String
	}
	confirmedAt, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return domain.ExperimentPreparation{}, false, fmt.Errorf("parse experiment preparation update time: %w", err)
	}
	preparation.LastConfirmedAt = confirmedAt.UTC()

	rows, err := s.db.QueryContext(ctx, "SELECT sequence_no, content FROM experiment_preparation_prompts WHERE experiment_id = ? ORDER BY sequence_no ASC", experimentID)
	if err != nil {
		return domain.ExperimentPreparation{}, false, fmt.Errorf("query experiment preparation prompts: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close experiment preparation prompt rows: %w", closeErr)
		}
	}()

	preparation.Prompts = make([]domain.ExperimentPreparationPrompt, 0)
	for rows.Next() {
		var prompt domain.ExperimentPreparationPrompt
		if err := rows.Scan(&prompt.SequenceNo, &prompt.Content); err != nil {
			return domain.ExperimentPreparation{}, false, fmt.Errorf("scan experiment preparation prompt: %w", err)
		}
		preparation.Prompts = append(preparation.Prompts, prompt)
	}
	// 単体テスト到達不可: SQLiteの完全に読み切ったRowsではrows.Errがnilとなり、Storeは固定のsqliteドライバを使用するため反復時エラーを注入できない。
	if err := rows.Err(); err != nil {
		return domain.ExperimentPreparation{}, false, fmt.Errorf("iterate experiment preparation prompts: %w", err)
	}

	return preparation, true, nil
}

// GetExperimentWorkspace は固定済み条件を含む実験ワークスペースを読み出す。
func (s *Store) GetExperimentWorkspace(ctx context.Context, experimentID string) (workspace domain.ExperimentWorkspace, found bool, err error) {
	workspace.ExperimentID = experimentID
	workspace.FixedConditions.ExperimentID = experimentID

	var hypothesis sql.NullString
	var updatedAt string
	var fixedAt string
	var operationFixedAt string
	err = s.db.QueryRowContext(ctx, "SELECT e.state, e.updated_at, c.id, c.purpose, c.hypothesis, c.environment_conditions, c.initial_input, c.evaluation_axes, c.fixed_at, o.operation_id, o.fixed_at FROM experiments e JOIN experiment_fixed_conditions c ON c.id = e.fixed_condition_id JOIN experiment_condition_fix_operations o ON o.fixed_condition_id = c.id WHERE e.id = ?", experimentID).Scan(&workspace.State, &updatedAt, &workspace.FixedConditions.FixedConditionID, &workspace.FixedConditions.Purpose, &hypothesis, &workspace.FixedConditions.EnvironmentConditions, &workspace.FixedConditions.InitialInput, &workspace.FixedConditions.EvaluationAxes, &fixedAt, &workspace.ConditionFixOperationID, &operationFixedAt)
	if err == sql.ErrNoRows {
		return domain.ExperimentWorkspace{}, false, nil
	}
	if err != nil {
		return domain.ExperimentWorkspace{}, false, fmt.Errorf("find experiment workspace: %w", err)
	}
	if hypothesis.Valid {
		workspace.FixedConditions.Hypothesis = &hypothesis.String
	}
	workspace.LastConfirmedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return domain.ExperimentWorkspace{}, false, fmt.Errorf("parse experiment workspace update time: %w", err)
	}
	workspace.FixedConditions.FixedAt, err = time.Parse(time.RFC3339Nano, fixedAt)
	if err != nil {
		return domain.ExperimentWorkspace{}, false, fmt.Errorf("parse experiment condition fixed time: %w", err)
	}
	workspace.ConditionFixOperationAt, err = time.Parse(time.RFC3339Nano, operationFixedAt)
	if err != nil {
		return domain.ExperimentWorkspace{}, false, fmt.Errorf("parse experiment condition operation fixed time: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, "SELECT sequence_no, content FROM experiment_fixed_condition_prompts WHERE fixed_condition_id = ? ORDER BY sequence_no ASC", workspace.FixedConditions.FixedConditionID)
	if err != nil {
		return domain.ExperimentWorkspace{}, false, fmt.Errorf("query experiment workspace prompts: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			// 単体テスト到達不可: database/sqlはNextがEOFを返した時点でdriver.Rows.Closeの失敗をrows.Errへ移すため、defer内だけでClose失敗を受け取る経路は作れない。
			err = fmt.Errorf("close experiment workspace prompt rows: %w", closeErr)
		}
	}()

	workspace.FixedConditions.Prompts = make([]domain.ExperimentPreparationPrompt, 0)
	for rows.Next() {
		var prompt domain.ExperimentPreparationPrompt
		if err := rows.Scan(&prompt.SequenceNo, &prompt.Content); err != nil {
			return domain.ExperimentWorkspace{}, false, fmt.Errorf("scan experiment workspace prompt: %w", err)
		}
		workspace.FixedConditions.Prompts = append(workspace.FixedConditions.Prompts, prompt)
	}
	// 単体テスト到達不可: SQLiteの完全に読み切ったRowsではrows.Errがnilとなり、Storeは固定のsqliteドライバを使用するため反復時エラーを注入できない。
	if err := rows.Err(); err != nil {
		return domain.ExperimentWorkspace{}, false, fmt.Errorf("iterate experiment workspace prompts: %w", err)
	}

	workspace.Runs, err = s.findExperimentWorkspaceRuns(ctx, experimentID)
	if err != nil {
		return domain.ExperimentWorkspace{}, false, err
	}
	workspace.Evaluations, err = s.findExperimentWorkspaceEvaluations(ctx, experimentID)
	if err != nil {
		return domain.ExperimentWorkspace{}, false, err
	}

	return workspace, true, nil
}

// findExperimentWorkspaceRuns は実験に属するrunの安全な進行状況を取得する。
func (s *Store) findExperimentWorkspaceRuns(ctx context.Context, experimentID string) ([]domain.ExperimentWorkspaceRun, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, state, summary, updated_at FROM experiment_runs WHERE experiment_id = ? ORDER BY created_at ASC, id ASC", experimentID)
	if err != nil {
		return nil, fmt.Errorf("query experiment workspace runs: %w", err)
	}
	defer rows.Close()

	runs := make([]domain.ExperimentWorkspaceRun, 0)
	for rows.Next() {
		var run domain.ExperimentWorkspaceRun
		var summary sql.NullString
		var updatedAt string
		if err := rows.Scan(&run.ID, &run.State, &summary, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan experiment workspace run: %w", err)
		}
		if summary.Valid {
			run.Summary = &summary.String
		}
		run.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse experiment workspace run update time: %w", err)
		}
		runs = append(runs, run)
	}
	// 単体テスト到達不可: SQLiteの完全に読み切ったRowsではrows.Errがnilとなり、Storeは固定のsqliteドライバを使用するため反復時エラーを注入できない。
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate experiment workspace runs: %w", err)
	}

	return runs, nil
}

// findExperimentWorkspaceEvaluations は実験に属するevaluationの安全な進行状況を取得する。
func (s *Store) findExperimentWorkspaceEvaluations(ctx context.Context, experimentID string) ([]domain.ExperimentWorkspaceEvaluation, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, state, summary, updated_at FROM experiment_evaluations WHERE experiment_id = ? ORDER BY created_at ASC, id ASC", experimentID)
	if err != nil {
		return nil, fmt.Errorf("query experiment workspace evaluations: %w", err)
	}
	defer rows.Close()

	evaluations := make([]domain.ExperimentWorkspaceEvaluation, 0)
	for rows.Next() {
		var evaluation domain.ExperimentWorkspaceEvaluation
		var summary sql.NullString
		var updatedAt string
		if err := rows.Scan(&evaluation.ID, &evaluation.State, &summary, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan experiment workspace evaluation: %w", err)
		}
		if summary.Valid {
			evaluation.Summary = &summary.String
		}
		evaluation.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse experiment workspace evaluation update time: %w", err)
		}
		evaluations = append(evaluations, evaluation)
	}
	// 単体テスト到達不可: SQLiteの完全に読み切ったRowsではrows.Errがnilとなり、Storeは固定のsqliteドライバを使用するため反復時エラーを注入できない。
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate experiment workspace evaluations: %w", err)
	}

	return evaluations, nil
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
	var candidatePrompts string
	err := s.db.QueryRowContext(ctx, "SELECT id, purpose, decision, hypothesis, candidate_prompts, evaluation_criteria, environment_conditions, initial_input, success_criteria, required_conditions, open_question, created_at FROM briefing_versions WHERE preparation_session_id = ? ORDER BY version_no DESC LIMIT 1", briefingSessionID).Scan(&brief.VersionID, &brief.Purpose, &brief.Decision, &hypothesis, &candidatePrompts, &brief.EvaluationCriteria, &brief.EnvironmentConditions, &brief.InitialInput, &brief.SuccessCriteria, &brief.RequiredConditions, &openQuestion, &createdAt)
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
	if err := json.Unmarshal([]byte(candidatePrompts), &brief.CandidatePrompts); err != nil {
		return nil, time.Time{}, fmt.Errorf("unmarshal candidate prompts: %w", err)
	}
	parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("parse briefing version creation time: %w", err)
	}

	return brief, parsedCreatedAt.UTC(), nil
}

// 採用済みブリーフ版からの準備中実験原子的保存。
func (s *Store) CreateExperimentFromBrief(ctx context.Context, requestID, briefingSessionID, briefVersionID string) (domain.ExperimentCreation, bool, error) {
	existing, found, err := s.findExperimentCreation(ctx, requestID)
	if err != nil {
		return domain.ExperimentCreation{}, false, err
	}
	if found {
		if existing.BriefingSessionID != briefingSessionID || existing.BriefVersionID != briefVersionID {
			return domain.ExperimentCreation{}, false, apperr.New(apperr.CodeExperimentCreateRequestConflict)
		}

		return domain.ExperimentCreation{ExperimentID: existing.ExperimentID, State: "preparing"}, false, nil
	}

	tx, err := s.beginBriefingTransaction(ctx)
	if err != nil {
		return domain.ExperimentCreation{}, false, fmt.Errorf("begin experiment creation: %w", err)
	}
	var sessionState string
	if err := tx.QueryRowContext(ctx, "SELECT state FROM preparation_sessions WHERE id = ? AND kind = ?", briefingSessionID, "experiment_brief").Scan(&sessionState); err != nil {
		_ = tx.Rollback()
		if err == sql.ErrNoRows {
			return domain.ExperimentCreation{}, false, apperr.New(apperr.CodeBriefingNotFound)
		}

		return domain.ExperimentCreation{}, false, fmt.Errorf("find experiment briefing adoption session: %w", err)
	}
	if sessionState != domain.BriefingStartStateStarted {
		_ = tx.Rollback()

		return domain.ExperimentCreation{}, false, apperr.New(apperr.CodeBriefingNotActive)
	}
	brief, err := findExperimentBriefForAdoption(ctx, tx, briefingSessionID, briefVersionID)
	if err != nil {
		_ = tx.Rollback()

		return domain.ExperimentCreation{}, false, err
	}
	if err := validateExperimentBriefForAdoption(brief); err != nil {
		_ = tx.Rollback()

		return domain.ExperimentCreation{}, false, err
	}
	experimentID, err := newBriefingIdentifier()
	if err != nil {
		_ = tx.Rollback()

		return domain.ExperimentCreation{}, false, fmt.Errorf("generate experiment identifier: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, "INSERT INTO experiments (id, purpose, state, updated_at) VALUES (?, ?, ?, ?)", experimentID, brief.Purpose, "preparing", now); err != nil {
		_ = tx.Rollback()

		return domain.ExperimentCreation{}, false, fmt.Errorf("insert experiment: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO experiment_preparations (experiment_id, briefing_session_id, briefing_version_id, hypothesis, environment_conditions, initial_input, evaluation_criteria, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", experimentID, briefingSessionID, briefVersionID, brief.Hypothesis, brief.EnvironmentConditions, brief.InitialInput, brief.EvaluationCriteria, now, now); err != nil {
		_ = tx.Rollback()

		return domain.ExperimentCreation{}, false, fmt.Errorf("insert experiment preparation: %w", err)
	}
	for index, prompt := range brief.CandidatePrompts {
		if _, err := tx.ExecContext(ctx, "INSERT INTO experiment_preparation_prompts (experiment_id, sequence_no, content) VALUES (?, ?, ?)", experimentID, index+1, prompt); err != nil {
			_ = tx.Rollback()

			return domain.ExperimentCreation{}, false, fmt.Errorf("insert experiment preparation prompt: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO experiment_creation_operations (request_id, briefing_session_id, briefing_version_id, experiment_id) VALUES (?, ?, ?, ?)", requestID, briefingSessionID, briefVersionID, experimentID); err != nil {
		_ = tx.Rollback()
		if isExperimentCreationRequestConflict(err) {
			existing, found, findErr := s.findExperimentCreation(ctx, requestID)
			if findErr != nil {
				return domain.ExperimentCreation{}, false, findErr
			}
			if found && existing.BriefingSessionID == briefingSessionID && existing.BriefVersionID == briefVersionID {
				return domain.ExperimentCreation{ExperimentID: existing.ExperimentID, State: "preparing"}, false, nil
			}
		}

		return domain.ExperimentCreation{}, false, fmt.Errorf("insert experiment creation operation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.ExperimentCreation{}, false, fmt.Errorf("commit experiment creation: %w", err)
	}

	return domain.ExperimentCreation{ExperimentID: experimentID, State: "preparing"}, true, nil
}

// experimentCreationOperation は採用操作の冪等性確認用記録。
type experimentCreationOperation struct {
	BriefingSessionID string
	BriefVersionID    string
	ExperimentID      string
}

// request ID対応採用操作取得。
func (s *Store) findExperimentCreation(ctx context.Context, requestID string) (experimentCreationOperation, bool, error) {
	var operation experimentCreationOperation
	err := s.db.QueryRowContext(ctx, "SELECT briefing_session_id, briefing_version_id, experiment_id FROM experiment_creation_operations WHERE request_id = ?", requestID).Scan(&operation.BriefingSessionID, &operation.BriefVersionID, &operation.ExperimentID)
	if err == sql.ErrNoRows {
		return experimentCreationOperation{}, false, nil
	}
	if err != nil {
		return experimentCreationOperation{}, false, fmt.Errorf("find experiment creation operation: %w", err)
	}

	return operation, true, nil
}

// 指定session所有ブリーフ版取得。
func findExperimentBriefForAdoption(ctx context.Context, tx briefingTransaction, briefingSessionID, briefVersionID string) (domain.ExperimentBrief, error) {
	var brief domain.ExperimentBrief
	var hypothesis sql.NullString
	var prompts string
	err := tx.QueryRowContext(ctx, "SELECT purpose, hypothesis, candidate_prompts, evaluation_criteria, environment_conditions, initial_input FROM briefing_versions WHERE id = ? AND preparation_session_id = ?", briefVersionID, briefingSessionID).Scan(&brief.Purpose, &hypothesis, &prompts, &brief.EvaluationCriteria, &brief.EnvironmentConditions, &brief.InitialInput)
	if err == sql.ErrNoRows {
		return domain.ExperimentBrief{}, apperr.New(apperr.CodeExperimentBriefVersionNotFound)
	}
	if err != nil {
		return domain.ExperimentBrief{}, fmt.Errorf("find experiment brief version: %w", err)
	}
	if hypothesis.Valid {
		brief.Hypothesis = &hypothesis.String
	}
	if err := json.Unmarshal([]byte(prompts), &brief.CandidatePrompts); err != nil {
		return domain.ExperimentBrief{}, fmt.Errorf("unmarshal experiment brief prompts: %w", err)
	}

	return brief, nil
}

// 実験作成用採用値検証。
func validateExperimentBriefForAdoption(brief domain.ExperimentBrief) error {
	if strings.TrimSpace(brief.Purpose) == "" || strings.TrimSpace(brief.EvaluationCriteria) == "" || strings.TrimSpace(brief.EnvironmentConditions) == "" || len(brief.CandidatePrompts) < 2 {
		return apperr.New(apperr.CodeExperimentBriefIncomplete)
	}
	for _, prompt := range brief.CandidatePrompts {
		if strings.TrimSpace(prompt) == "" {
			return apperr.New(apperr.CodeExperimentBriefIncomplete)
		}
	}

	return nil
}

// 採用request ID一意制約競合判定。
func isExperimentCreationRequestConflict(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed: experiment_creation_operations.request_id")
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

// isBriefingStopRequestConflict は停止request IDの一意制約競合を判定。
func isBriefingStopRequestConflict(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed: briefing_stop_operations.request_id")
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
	s.listMu.Lock()
	defer s.listMu.Unlock()

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
