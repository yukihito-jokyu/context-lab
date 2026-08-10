package domain

import "testing"

// 実験準備必須項目の充足状態を検証。
func TestExperimentPreparationRequiredFields(t *testing.T) {
	tests := []struct {
		name        string
		preparation ExperimentPreparation
		want        ExperimentPreparationRequiredFields
	}{
		{
			name: "すべての必須項目が充足している",
			preparation: ExperimentPreparation{
				Purpose:               "目的",
				EnvironmentConditions: "隔離環境",
				InitialInput:          "初期入力",
				Prompts: []ExperimentPreparationPrompt{
					{
						Content: "prompt A",
					},
					{
						Content: "prompt B",
					},
				},
				EvaluationAxes: "正確性",
			},
			want: ExperimentPreparationRequiredFields{
				Purpose:               true,
				EnvironmentConditions: true,
				InitialInput:          true,
				Prompts:               true,
				EvaluationAxes:        true,
			},
		},
		{
			name: "空欄と二件未満のpromptを未充足にする",
			preparation: ExperimentPreparation{
				Prompts: []ExperimentPreparationPrompt{
					{
						Content: " ",
					},
				},
			},
			want: ExperimentPreparationRequiredFields{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.preparation.RequiredFields(); got != tt.want {
				t.Errorf("RequiredFields() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// 固定条件の必須入力検証。
func TestExperimentFixedConditionsValid(t *testing.T) {
	tests := []struct {
		name       string
		conditions ExperimentFixedConditions
		want       bool
	}{
		{
			name: "全条件が揃う",
			conditions: ExperimentFixedConditions{
				Purpose:               "目的",
				EnvironmentConditions: "環境",
				InitialInput:          "入力",
				Prompts: []ExperimentPreparationPrompt{
					{
						SequenceNo: 1,
						Content:    "prompt",
					},
				},
				EvaluationAxes: "評価",
			},
			want: true,
		},
		{
			name:       "必須条件不足",
			conditions: ExperimentFixedConditions{},
			want:       false,
		},
		{
			name: "空白prompt",
			conditions: ExperimentFixedConditions{
				Purpose:               "目的",
				EnvironmentConditions: "環境",
				InitialInput:          "入力",
				Prompts: []ExperimentPreparationPrompt{
					{
						SequenceNo: 1,
						Content:    " ",
					},
				},
				EvaluationAxes: "評価",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.conditions.Valid(); got != tt.want {
				t.Errorf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}
