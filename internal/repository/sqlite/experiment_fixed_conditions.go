package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// FixExperimentConditions は現在の下書きと一致する条件を不変artifactとして固定する。
func (s *Store) FixExperimentConditions(ctx context.Context, conditions domain.ExperimentFixedConditions) (domain.ExperimentFixedConditions, error) {
	var lastErr error
	for attempt := range 8 {
		fixed, err := s.fixExperimentConditions(ctx, conditions)
		if !isSQLiteBusy(err) {
			return fixed, err
		}
		lastErr = err
		if err := waitDraftSaveRetry(ctx, time.Millisecond<<attempt); err != nil {
			return domain.ExperimentFixedConditions{}, err
		}
	}

	return domain.ExperimentFixedConditions{}, lastErr
}

// fixExperimentConditions は一回のSQLite transactionで条件を固定する。
func (s *Store) fixExperimentConditions(ctx context.Context, conditions domain.ExperimentFixedConditions) (domain.ExperimentFixedConditions, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.ExperimentFixedConditions{}, fmt.Errorf("begin condition fix transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	existing, found, err := findExperimentConditionFixOperation(ctx, tx, conditions.RequestID)
	if err != nil {
		return domain.ExperimentFixedConditions{}, err
	}
	if found {
		if existing.ExperimentID != conditions.ExperimentID {
			return domain.ExperimentFixedConditions{}, apperr.New(apperr.CodeFixConditionsRequestInvalid)
		}

		return existing, nil
	}

	var state string
	var fixedConditionID sql.NullString
	err = tx.QueryRowContext(ctx, "SELECT state, fixed_condition_id FROM experiments WHERE id = ?", conditions.ExperimentID).Scan(&state, &fixedConditionID)
	if err == sql.ErrNoRows {
		return domain.ExperimentFixedConditions{}, apperr.New(apperr.CodeExperimentPreparationNotFound)
	}
	if err != nil {
		return domain.ExperimentFixedConditions{}, fmt.Errorf("find condition fix experiment: %w", err)
	}
	if fixedConditionID.Valid || state == "ready" {
		return domain.ExperimentFixedConditions{}, apperr.New(apperr.CodeExperimentConditionsAlreadyFixed)
	}
	if state != "preparing" {
		return domain.ExperimentFixedConditions{}, apperr.New(apperr.CodeExperimentConditionsAlreadyFixed)
	}

	draft, err := findCurrentExperimentPreparationDraft(ctx, tx, conditions.ExperimentID)
	if err != nil {
		return domain.ExperimentFixedConditions{}, err
	}
	if !sameExperimentConditions(draft, conditions) {
		return domain.ExperimentFixedConditions{}, apperr.New(apperr.CodeExperimentConditionsConflict)
	}

	conditions.FixedConditionID, err = newBriefingIdentifier()
	if err != nil {
		return domain.ExperimentFixedConditions{}, fmt.Errorf("generate fixed condition ID: %w", err)
	}
	conditions.OperationID, err = newBriefingIdentifier()
	if err != nil {
		return domain.ExperimentFixedConditions{}, fmt.Errorf("generate condition fix operation ID: %w", err)
	}
	conditions.FixedAt = time.Now().UTC()
	payload, err := json.Marshal(struct {
		Purpose               string                               `json:"purpose"`
		Hypothesis            *string                              `json:"hypothesis,omitempty"`
		EnvironmentConditions string                               `json:"environmentConditions"`
		InitialInput          string                               `json:"initialInput"`
		Prompts               []domain.ExperimentPreparationPrompt `json:"prompts"`
		EvaluationAxes        string                               `json:"evaluationAxes"`
	}{conditions.Purpose, conditions.Hypothesis, conditions.EnvironmentConditions, conditions.InitialInput, conditions.Prompts, conditions.EvaluationAxes})
	// 単体テスト到達不可: artifact payloadはstring、*string、int、stringだけで構成され、encoding/jsonのMarshalで失敗し得る値をdomain.ExperimentFixedConditionsは保持できない。
	if err != nil {
		return domain.ExperimentFixedConditions{}, fmt.Errorf("marshal fixed condition artifact: %w", err)
	}
	fixedAt := conditions.FixedAt.Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, "INSERT INTO experiment_fixed_conditions (id, experiment_id, purpose, hypothesis, environment_conditions, initial_input, evaluation_axes, artifact_payload, fixed_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", conditions.FixedConditionID, conditions.ExperimentID, conditions.Purpose, conditions.Hypothesis, conditions.EnvironmentConditions, conditions.InitialInput, conditions.EvaluationAxes, string(payload), fixedAt); err != nil {
		return domain.ExperimentFixedConditions{}, fmt.Errorf("insert fixed conditions: %w", err)
	}
	for _, prompt := range conditions.Prompts {
		if _, err := tx.ExecContext(ctx, "INSERT INTO experiment_fixed_condition_prompts (fixed_condition_id, sequence_no, content) VALUES (?, ?, ?)", conditions.FixedConditionID, prompt.SequenceNo, prompt.Content); err != nil {
			return domain.ExperimentFixedConditions{}, fmt.Errorf("insert fixed condition prompt: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, "UPDATE experiments SET state = ?, fixed_condition_id = ?, updated_at = ? WHERE id = ?", "ready", conditions.FixedConditionID, fixedAt, conditions.ExperimentID); err != nil {
		return domain.ExperimentFixedConditions{}, fmt.Errorf("transition experiment conditions: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO experiment_condition_fix_operations (request_id, experiment_id, fixed_condition_id, operation_id, fixed_at) VALUES (?, ?, ?, ?, ?)", conditions.RequestID, conditions.ExperimentID, conditions.FixedConditionID, conditions.OperationID, fixedAt); err != nil {
		// 単体テスト到達不可: modernc SQLiteは同一書込transactionをbusyで直列化するため、事前読込後に別接続が同じrequest IDをcommitする競合は外側のbusy再試行で解消される。分岐は別driver設定に対する防御処理。
		if isConditionFixRequestConflict(err) {
			// 単体テスト到達不可: Storeが保持する*sql.TxのRollback結果は差し替え不能で、正常なSQLite transactionのRollbackは必ずnilまたはsql.ErrTxDoneを返す。
			if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
				return domain.ExperimentFixedConditions{}, fmt.Errorf("rollback condition fix transaction after request conflict: %w", rollbackErr)
			}
			// 単体テスト到達不可: 同一request IDの書込競合がmodernc SQLiteでは上記のbusy再試行に置換されるため、この再読込はdriver固有の防御経路である。
			existing, found, findErr := findExperimentConditionFixOperation(ctx, s.db, conditions.RequestID)
			// 単体テスト到達不可: 同一request IDの書込競合がmodernc SQLiteでは上記のbusy再試行に置換されるため、この再読込失敗はdriver固有の防御経路である。
			if findErr != nil {
				return domain.ExperimentFixedConditions{}, findErr
			}
			// 単体テスト到達不可: 同一request IDの書込競合がmodernc SQLiteでは上記のbusy再試行に置換されるため、再読込snapshot返却はdriver固有の防御経路である。
			if found && existing.ExperimentID == conditions.ExperimentID {
				return existing, nil
			}
			// 単体テスト到達不可: 同一request IDの書込競合がmodernc SQLiteでは上記のbusy再試行に置換されるため、異なる実験のsnapshot拒否はdriver固有の防御経路である。
			if found {
				return domain.ExperimentFixedConditions{}, apperr.New(apperr.CodeFixConditionsRequestInvalid)
			}
		}

		return domain.ExperimentFixedConditions{}, fmt.Errorf("insert condition fix operation: %w", err)
	}
	// 単体テスト到達不可: Storeは実SQL transactionのみを保持し、SQLiteのcommit I/O障害をtest doubleで注入する境界を持たない。
	if err := tx.Commit(); err != nil {
		return domain.ExperimentFixedConditions{}, fmt.Errorf("commit condition fix transaction: %w", err)
	}

	return conditions, nil
}

// findCurrentExperimentPreparationDraft は実験に現在保存された下書きを取得する。
func findCurrentExperimentPreparationDraft(ctx context.Context, queryer conditionFixQueryer, experimentID string) (domain.ExperimentFixedConditions, error) {
	var draft domain.ExperimentFixedConditions
	var hypothesis sql.NullString
	err := queryer.QueryRowContext(ctx, "SELECT e.purpose, p.hypothesis, p.environment_conditions, p.initial_input, p.evaluation_criteria FROM experiments e JOIN experiment_preparations p ON p.experiment_id = e.id WHERE e.id = ?", experimentID).Scan(&draft.Purpose, &hypothesis, &draft.EnvironmentConditions, &draft.InitialInput, &draft.EvaluationAxes)
	if err == sql.ErrNoRows {
		return domain.ExperimentFixedConditions{}, apperr.New(apperr.CodeExperimentPreparationNotFound)
	}
	if err != nil {
		return domain.ExperimentFixedConditions{}, fmt.Errorf("find current preparation draft: %w", err)
	}
	if hypothesis.Valid {
		draft.Hypothesis = &hypothesis.String
	}
	rows, err := queryer.QueryContext(ctx, "SELECT sequence_no, content FROM experiment_preparation_prompts WHERE experiment_id = ? ORDER BY sequence_no", experimentID)
	if err != nil {
		return domain.ExperimentFixedConditions{}, fmt.Errorf("find current preparation prompts: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		var prompt domain.ExperimentPreparationPrompt
		if err := rows.Scan(&prompt.SequenceNo, &prompt.Content); err != nil {
			return domain.ExperimentFixedConditions{}, fmt.Errorf("scan current preparation prompt: %w", err)
		}
		draft.Prompts = append(draft.Prompts, prompt)
	}
	// 単体テスト到達不可: conditionFixQueryerはdatabase/sqlの*sql.Rowsだけを返すため、modernc SQLiteの反復中driver errorをtest doubleで注入できない。
	if err := rows.Err(); err != nil {
		return domain.ExperimentFixedConditions{}, fmt.Errorf("iterate current preparation prompts: %w", err)
	}

	return draft, nil
}

// findExperimentConditionFixOperation はrequest IDに対応する固定結果snapshotを取得する。
func findExperimentConditionFixOperation(ctx context.Context, queryer conditionFixQueryer, requestID string) (domain.ExperimentFixedConditions, bool, error) {
	var conditions domain.ExperimentFixedConditions
	var fixedAt string
	var hypothesis sql.NullString
	err := queryer.QueryRowContext(ctx, "SELECT o.experiment_id, o.fixed_condition_id, o.operation_id, o.fixed_at, c.purpose, c.hypothesis, c.environment_conditions, c.initial_input, c.evaluation_axes FROM experiment_condition_fix_operations o JOIN experiment_fixed_conditions c ON c.id = o.fixed_condition_id WHERE o.request_id = ?", requestID).Scan(&conditions.ExperimentID, &conditions.FixedConditionID, &conditions.OperationID, &fixedAt, &conditions.Purpose, &hypothesis, &conditions.EnvironmentConditions, &conditions.InitialInput, &conditions.EvaluationAxes)
	if err == sql.ErrNoRows {
		return domain.ExperimentFixedConditions{}, false, nil
	}
	if err != nil {
		return domain.ExperimentFixedConditions{}, false, fmt.Errorf("find condition fix operation: %w", err)
	}
	fixedAtValue, err := time.Parse(time.RFC3339Nano, fixedAt)
	if err != nil {
		return domain.ExperimentFixedConditions{}, false, fmt.Errorf("parse condition fixed time: %w", err)
	}
	conditions.RequestID = requestID
	conditions.FixedAt = fixedAtValue.UTC()
	if hypothesis.Valid {
		conditions.Hypothesis = &hypothesis.String
	}
	rows, err := queryer.QueryContext(ctx, "SELECT sequence_no, content FROM experiment_fixed_condition_prompts WHERE fixed_condition_id = ? ORDER BY sequence_no", conditions.FixedConditionID)
	if err != nil {
		return domain.ExperimentFixedConditions{}, false, fmt.Errorf("find fixed condition prompts: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		var prompt domain.ExperimentPreparationPrompt
		if err := rows.Scan(&prompt.SequenceNo, &prompt.Content); err != nil {
			return domain.ExperimentFixedConditions{}, false, fmt.Errorf("scan fixed condition prompt: %w", err)
		}
		conditions.Prompts = append(conditions.Prompts, prompt)
	}
	// 単体テスト到達不可: conditionFixQueryerはdatabase/sqlの*sql.Rowsだけを返すため、modernc SQLiteの反復中driver errorをtest doubleで注入できない。
	if err := rows.Err(); err != nil {
		return domain.ExperimentFixedConditions{}, false, fmt.Errorf("iterate fixed condition prompts: %w", err)
	}

	return conditions, true, nil
}

// sameExperimentConditions は下書きと固定要求が完全に一致するかを返す。
func sameExperimentConditions(draft, conditions domain.ExperimentFixedConditions) bool {
	if draft.Purpose != conditions.Purpose || draft.EnvironmentConditions != conditions.EnvironmentConditions || draft.InitialInput != conditions.InitialInput || draft.EvaluationAxes != conditions.EvaluationAxes || !sameOptionalString(draft.Hypothesis, conditions.Hypothesis) || len(draft.Prompts) != len(conditions.Prompts) {
		return false
	}
	for index, prompt := range draft.Prompts {
		if prompt.SequenceNo != conditions.Prompts[index].SequenceNo || prompt.Content != conditions.Prompts[index].Content {
			return false
		}
	}

	return true
}

// sameOptionalString はnullable文字列が一致するかを返す。
func sameOptionalString(first, second *string) bool {
	if first == nil || second == nil {
		return first == nil && second == nil
	}

	return *first == *second
}

// conditionFixQueryer は固定条件の読込境界。
type conditionFixQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

// isConditionFixRequestConflict は固定request IDの一意制約競合を判定。
func isConditionFixRequestConflict(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed: experiment_condition_fix_operations.request_id")
}
