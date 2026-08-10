package wails

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
	"github.com/yukihito-jokyu/context-lab/internal/usecase"
)

// 派生元queryの安全DTO。
func TestExperimentDerivationSourcesHandlerGetDerivationSource(t *testing.T) {
	when := time.Date(2026, time.August, 10, 9, 0, 0, 0, time.FixedZone("JST", 9*60*60))
	tests := []struct {
		name           string
		reader         handlerDerivationSourceReader
		wantCode       apperr.Code
		wantConclusion bool
	}{
		{
			name: "固定条件と結論をUTCで返す",
			reader: handlerDerivationSourceReader{
				found: true,
				source: domain.ExperimentDerivationSource{
					ExperimentID:     "experiment-1",
					Purpose:          "目的",
					CanCreateDerived: true,
					FixedConditions: &domain.ExperimentFixedConditions{
						FixedConditionID: "fixed-1",
						Purpose:          "固定目的",
						Prompts: []domain.ExperimentPreparationPrompt{
							{
								SequenceNo: 2,
								Content:    "second",
							},
						},
						FixedAt: when,
					},
					Conclusion: &domain.ExperimentConclusion{
						ConclusionID: "conclusion-1",
						Conclusion:   "結論",
						State:        "finalized",
						FinalizedAt:  when,
					},
				},
			},
			wantConclusion: true,
		},
		{
			name:     "内部エラーを安全に返す",
			reader:   handlerDerivationSourceReader{err: errors.New("private database detail")},
			wantCode: apperr.CodeExperimentDerivationSourceUnavailable,
		},
		{
			name: "未確定結論を返さない",
			reader: handlerDerivationSourceReader{
				found: true,
				source: domain.ExperimentDerivationSource{
					ExperimentID: "experiment-1",
					FixedConditions: &domain.ExperimentFixedConditions{
						FixedConditionID: "fixed-1",
						FixedAt:          when,
					},
					Conclusion: &domain.ExperimentConclusion{
						ConclusionID: "conclusion-1",
						Conclusion:   "下書き",
						State:        "draft",
						FinalizedAt:  when,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewExperimentDerivationSourcesHandler(usecase.NewGetDerivationSource(&tt.reader), newTestLogger()).GetDerivationSource("experiment-1")
			if tt.wantCode != "" {
				if got.Error == nil || got.Error.Code != string(tt.wantCode) {
					t.Errorf("Error = %+v, want code %q", got.Error, tt.wantCode)
				}
				return
			}
			if got.Data == nil || got.Data.Source.FixedConditions == nil {
				t.Fatal("Data = nil or incomplete, want derivation source")
			}
			if (got.Data.Source.Conclusion != nil) != tt.wantConclusion {
				t.Errorf("Conclusion = %+v, want present %v", got.Data.Source.Conclusion, tt.wantConclusion)
			}
			if !tt.wantConclusion {
				return
			}
			if got.Data.Source.FixedConditions.FixedAt.Location() != time.UTC || got.Data.Source.Conclusion.FinalizedAt.Location() != time.UTC {
				t.Errorf("times = %+v, want UTC", got.Data)
			}
		})
	}
}

// 派生元queryの依存欠落と未知エラー。
func TestExperimentDerivationSourcesHandlerFailure(t *testing.T) {
	tests := []struct {
		name string
		got  GetDerivationSourceResponse
		want apperr.Code
	}{
		{
			name: "依存欠落",
			got:  NewExperimentDerivationSourcesHandler(nil, newTestLogger()).GetDerivationSource("experiment-1"),
			want: apperr.CodeExperimentDerivationSourceUnavailable,
		},
		{
			name: "未知エラー",
			got:  NewExperimentDerivationSourcesHandler(nil, newTestLogger()).fail(context.Background(), errors.New("private credential")),
			want: apperr.CodeUnexpected,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got.Error == nil || tt.got.Error.Code != string(tt.want) {
				t.Errorf("Error = %+v, want code %q", tt.got.Error, tt.want)
			}
		})
	}
}

// handlerDerivationSourceReader は派生元readerのtest double。
type handlerDerivationSourceReader struct {
	source domain.ExperimentDerivationSource
	found  bool
	err    error
}

// GetDerivationSource は設定済み派生元を返す。
func (r *handlerDerivationSourceReader) GetDerivationSource(context.Context, string) (domain.ExperimentDerivationSource, bool, error) {
	return r.source, r.found, r.err
}
