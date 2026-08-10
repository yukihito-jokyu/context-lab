package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
)

// SQLite評価詳細は根拠、結果、失敗理由および照合状態を同じsnapshotから再読込する。
func TestStoreGetEvaluationDetail(t *testing.T) {
	store, evaluationID := evaluationDetailStore(t)
	confirmedAt := time.Date(2026, time.August, 10, 1, 2, 3, 0, time.UTC)
	confirmedValue := confirmedAt.Format(time.RFC3339Nano)
	if _, err := store.db.ExecContext(context.Background(), `UPDATE experiment_evaluations SET state = ?, summary = ?, result_status = ?, result_reason_code = ?, last_observed_at = ?, reconciliation_state = ?, created_at = ?, updated_at = ? WHERE id = ?`, domain.ExperimentEvaluationStateFailed, "評価要約", "partial", "EVALUATION_TIMEOUT", confirmedValue, "confirmed", confirmedValue, confirmedValue, evaluationID); err != nil {
		t.Fatalf("update evaluation detail fixture: %v", err)
	}
	if _, err := store.db.ExecContext(context.Background(), "UPDATE experiment_evaluation_operations SET state = ?, updated_at = ? WHERE evaluation_id = ?", domain.ExperimentEvaluationStateFailed, confirmedValue, evaluationID); err != nil {
		t.Fatalf("update evaluation operation fixture: %v", err)
	}

	detail, found, err := store.GetEvaluationDetail(context.Background(), evaluationID)
	if err != nil {
		t.Fatalf("GetEvaluationDetail() error = %v", err)
	}
	if !found {
		t.Fatal("GetEvaluationDetail() found = false, want true")
	}
	if detail.Evaluation.ID != evaluationID || detail.Evidence.RunSummary != "安全なrun要約" || detail.Evidence.EvaluationAxes == "" {
		t.Errorf("detail = %+v, want persisted safe facts", detail)
	}
	if detail.Result.Status != "partial" || detail.Result.Summary == nil || *detail.Result.Summary != "評価要約" || detail.Result.ReasonCode != "EVALUATION_TIMEOUT" {
		t.Errorf("result = %+v, want persisted partial result", detail.Result)
	}
	if detail.Failure == nil || detail.Failure.Code != "EVALUATION_TIMEOUT" || !detail.Reconciliation.LastObservedAt.Equal(confirmedAt) || !detail.LastConfirmedAt.Equal(confirmedAt) {
		t.Errorf("detail = %+v, want failure and confirmed times", detail)
	}
}

// SQLite評価詳細は未検出、SQL失敗および不正時刻を呼出側へ明示する。
func TestStoreGetEvaluationDetailErrors(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*testing.T, *Store, string)
		queryID   string
		wantFound bool
		wantError bool
	}{
		{
			name:      "未検出",
			setup:     func(_ *testing.T, _ *Store, _ string) {},
			queryID:   "missing-evaluation",
			wantFound: false,
		},
		{
			name: "評価作成時刻が不正",
			setup: func(t *testing.T, s *Store, id string) {
				setEvaluationDetailValue(t, s, "created_at", "invalid", id)
			},
			wantError: true,
		},
		{
			name: "評価更新時刻が不正",
			setup: func(t *testing.T, s *Store, id string) {
				setEvaluationDetailValue(t, s, "updated_at", "invalid", id)
			},
			wantError: true,
		},
		{
			name: "操作更新時刻が不正",
			setup: func(t *testing.T, s *Store, id string) {
				if _, err := s.db.ExecContext(context.Background(), "UPDATE experiment_evaluation_operations SET updated_at = ? WHERE evaluation_id = ?", "invalid", id); err != nil {
					t.Fatalf("update operation time: %v", err)
				}
			},
			wantError: true,
		},
		{
			name: "最終観測時刻が不正",
			setup: func(t *testing.T, s *Store, id string) {
				setEvaluationDetailValue(t, s, "last_observed_at", "invalid", id)
			},
			wantError: true,
		},
		{
			name: "SQL読込に失敗",
			setup: func(t *testing.T, s *Store, _ string) {
				if err := s.Close(); err != nil {
					t.Fatalf("Close() error = %v", err)
				}
			},
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, evaluationID := evaluationDetailStore(t)
			t.Cleanup(func() { _ = store.Close() })
			tt.setup(t, store, evaluationID)
			queryID := evaluationID
			if tt.queryID != "" {
				queryID = tt.queryID
			}
			_, found, err := store.GetEvaluationDetail(context.Background(), queryID)
			if found != tt.wantFound {
				t.Errorf("GetEvaluationDetail() found = %v, want %v", found, tt.wantFound)
			}
			if !tt.wantError {
				if err != nil {
					t.Errorf("GetEvaluationDetail() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Error("GetEvaluationDetail() error = nil, want read or parse error")
			}
		})
	}
}

// evaluationDetailStore は評価詳細を再読込できる最小のSQLite fixtureを生成する。
func evaluationDetailStore(t *testing.T) (*Store, string) {
	t.Helper()
	store, fixed := fixedExperimentPreparationStore(t)
	t.Cleanup(func() { _ = store.Close() })
	start, _, err := store.BeginExperiment(context.Background(), "evaluation-detail-start", fixed.ExperimentID)
	if err != nil {
		t.Fatalf("BeginExperiment() error = %v", err)
	}
	if err := store.CompleteExperimentRun(context.Background(), start.Runs[0].ID, "安全なrun要約"); err != nil {
		t.Fatalf("CompleteExperimentRun() error = %v", err)
	}
	evaluation, _, err := store.BeginRunEvaluation(context.Background(), "evaluation-detail-request", start.Runs[0].ID)
	if err != nil {
		t.Fatalf("BeginRunEvaluation() error = %v", err)
	}

	return store, evaluation.EvaluationID
}

// setEvaluationDetailValue は評価詳細fixtureの時刻列を更新する。
func setEvaluationDetailValue(t *testing.T, store *Store, column, value, evaluationID string) {
	t.Helper()
	if _, err := store.db.ExecContext(context.Background(), "UPDATE experiment_evaluations SET "+column+" = ? WHERE id = ?", value, evaluationID); err != nil {
		t.Fatalf("update evaluation %s: %v", column, err)
	}
}
