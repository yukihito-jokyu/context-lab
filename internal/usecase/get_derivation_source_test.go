package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	apperr "github.com/yukihito-jokyu/context-lab/internal/errors"
)

func TestGetDerivationSourceExecute(t *testing.T) {
	tests := []struct {
		name, id string
		found    bool
		err      error
		want     apperr.Code
	}{
		{
			name: "空入力",
			want: apperr.CodeExperimentDerivationSourceInvalid,
		},
		{
			name: "未検出",
			id:   "missing",
			want: apperr.CodeExperimentDerivationSourceNotFound,
		},
		{
			name:  "成功",
			id:    "experiment",
			found: true,
		},
		{
			name: "アプリケーションエラー",
			id:   "experiment",
			err:  apperr.New(apperr.CodeExperimentDerivationSourceNotFound),
			want: apperr.CodeExperimentDerivationSourceNotFound,
		},
		{
			name: "予期しないエラー",
			id:   "experiment",
			err:  errors.New("database unavailable"),
			want: apperr.CodeExperimentDerivationSourceUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := derivationReader{
				found: tt.found,
				err:   tt.err,
			}
			got, err := NewGetDerivationSource(reader).Execute(context.Background(), tt.id)
			if tt.want != "" {
				if !apperr.IsCode(err, tt.want) {
					t.Errorf("Execute() error=%v", err)
				}
				return
			}
			if err != nil || got.ExperimentID != "experiment" {
				t.Errorf("Execute()=(%+v,%v)", got, err)
			}
		})
	}
}

type derivationReader struct {
	found bool
	err   error
}

func (r derivationReader) GetDerivationSource(context.Context, string) (domain.ExperimentDerivationSource, bool, error) {
	return domain.ExperimentDerivationSource{ExperimentID: "experiment"}, r.found, r.err
}
