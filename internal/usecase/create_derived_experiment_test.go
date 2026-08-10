package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

func TestCreateDerivedExperimentExecute(t *testing.T) {
	value := "purpose"
	cases := []struct {
		name                    string
		request, source, reason string
		changes                 domain.DerivedExperimentChanges
		err                     error
		want                    apperr.Code
	}{
		{
			name: "invalid",
			want: apperr.CodeDerivedExperimentInvalid,
		},
		{
			name:    "known",
			request: "r",
			source:  "s",
			reason:  "x",
			changes: domain.DerivedExperimentChanges{
				Purpose: &value,
			},
			err:  apperr.New(apperr.CodeDerivedExperimentSourceNotEligible),
			want: apperr.CodeDerivedExperimentSourceNotEligible,
		},
		{
			name:    "unknown",
			request: "r",
			source:  "s",
			reason:  "x",
			changes: domain.DerivedExperimentChanges{
				Purpose: &value,
			},
			err:  errors.New("private"),
			want: apperr.CodeDerivedExperimentUnavailable,
		},
		{
			name:    "success",
			request: " r ",
			source:  " s ",
			reason:  " x ",
			changes: domain.DerivedExperimentChanges{
				Purpose: &value,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewCreateDerivedExperiment(derivedCreator{err: tc.err}).Execute(context.Background(), tc.request, tc.source, tc.changes, tc.reason)
			if tc.want != "" {
				if !apperr.IsCode(err, tc.want) {
					t.Errorf("error=%v", err)
				}
				return
			}
			if err != nil || got.ExperimentID == "" {
				t.Errorf("result=%+v error=%v", got, err)
			}
		})
	}
}

type derivedCreator struct{ err error }

func (c derivedCreator) CreateDerivedExperiment(context.Context, string, string, domain.DerivedExperimentChanges, string) (domain.DerivedExperiment, bool, error) {
	return domain.DerivedExperiment{ExperimentID: "derived"}, true, c.err
}
