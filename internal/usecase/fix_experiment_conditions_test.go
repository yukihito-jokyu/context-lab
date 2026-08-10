package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

// 条件固定commandの入力検証とstore委譲。
func TestFixExperimentConditionsExecute(t *testing.T) {
	tests := []struct {
		name       string
		conditions domain.ExperimentFixedConditions
		storeErr   error
		wantCode   apperr.Code
		wantCall   bool
	}{
		{
			name:       "条件を固定する",
			conditions: validFixedConditions(),
			wantCall:   true,
		},
		{
			name:       "request ID不足を拒否する",
			conditions: domain.ExperimentFixedConditions{ExperimentID: "experiment-1"},
			wantCode:   apperr.CodeFixConditionsRequestInvalid,
		},
		{
			name: "必須条件不足を拒否する",
			conditions: domain.ExperimentFixedConditions{
				RequestID:    "request-1",
				ExperimentID: "experiment-1",
			},
			wantCode: apperr.CodeConditionsInvalid,
		},
		{
			name:       "store失敗を安全に変換する",
			conditions: validFixedConditions(),
			storeErr:   errors.New("database detail"),
			wantCode:   apperr.CodeFixConditionsSaveFailed,
			wantCall:   true,
		},
		{
			name:       "アプリケーションエラーを維持する",
			conditions: validFixedConditions(),
			storeErr:   apperr.New(apperr.CodeExperimentConditionsConflict),
			wantCode:   apperr.CodeExperimentConditionsConflict,
			wantCall:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeExperimentConditionsStore{err: tt.storeErr}
			got, err := NewFixExperimentConditions(store).Execute(context.Background(), tt.conditions)
			if tt.wantCode != "" {
				if !apperr.IsCode(err, tt.wantCode) {
					t.Errorf("Execute() error = %v, want code %q", err, tt.wantCode)
				}
			} else if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if gotCall := store.called; gotCall != tt.wantCall {
				t.Errorf("FixExperimentConditions() called = %v, want %v", gotCall, tt.wantCall)
			}
			if tt.wantCode == "" && got.ExperimentID != tt.conditions.ExperimentID {
				t.Errorf("ExperimentID = %q, want %q", got.ExperimentID, tt.conditions.ExperimentID)
			}
		})
	}
}

// validFixedConditions は有効な固定条件を返す。
func validFixedConditions() domain.ExperimentFixedConditions {
	return domain.ExperimentFixedConditions{
		RequestID:             "request-1",
		ExperimentID:          "experiment-1",
		Purpose:               "目的",
		EnvironmentConditions: "環境",
		InitialInput:          "入力",
		Prompts: []domain.ExperimentPreparationPrompt{
			{
				SequenceNo: 1,
				Content:    "prompt",
			},
		},
		EvaluationAxes: "評価",
	}
}

// fakeExperimentConditionsStore は条件固定portのtest double。
type fakeExperimentConditionsStore struct {
	called bool
	err    error
}

// FixExperimentConditions は指定済み条件または失敗を返却。
func (s *fakeExperimentConditionsStore) FixExperimentConditions(_ context.Context, conditions domain.ExperimentFixedConditions) (domain.ExperimentFixedConditions, error) {
	s.called = true
	if s.err != nil {
		return domain.ExperimentFixedConditions{}, s.err
	}

	return conditions, nil
}
