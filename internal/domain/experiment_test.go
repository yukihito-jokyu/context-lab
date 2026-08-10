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
