package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
)

// 派生元queryのSQLite境界。
func TestStoreGetDerivationSource(t *testing.T) {
	tests := []struct {
		name, stage  string
		wantFound    bool
		wantErr      bool
		wantEligible bool
		wantReason   string
	}{
		{
			name:  "存在しない",
			stage: "missing",
		},
		{
			name:       "固定条件なし",
			stage:      "no-fixed",
			wantFound:  true,
			wantReason: "CONDITIONS_NOT_FIXED",
		},
		{
			name:       "結論なし",
			stage:      "no-conclusion",
			wantFound:  true,
			wantReason: "CONCLUSION_NOT_FINALIZED",
		},
		{
			name:       "非finalized結論",
			stage:      "draft",
			wantFound:  true,
			wantReason: "CONCLUSION_NOT_FINALIZED",
		},
		{
			name:         "NULL仮説とprompt順",
			stage:        "success",
			wantFound:    true,
			wantEligible: true,
		},
		{
			name:         "仮説を返す",
			stage:        "hypothesis",
			wantFound:    true,
			wantEligible: true,
		},
		{
			name:    "主query失敗",
			stage:   "query-error",
			wantErr: true,
		},
		{
			name:    "主scan失敗",
			stage:   "scan-error",
			wantErr: true,
		},
		{
			name:    "固定時刻不正",
			stage:   "fixed-time",
			wantErr: true,
		},
		{
			name:    "結論時刻不正",
			stage:   "conclusion-time",
			wantErr: true,
		},
		{
			name:    "prompt query失敗",
			stage:   "prompt-query",
			wantErr: true,
		},
		{
			name:    "prompt scan失敗",
			stage:   "prompt-scan",
			wantErr: true,
		},
		{
			name:    "prompt rows失敗",
			stage:   "prompt-rows",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found, err := newDerivationSourceStore(t, tt.stage).GetDerivationSource(context.Background(), "experiment")
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if found != tt.wantFound {
				t.Errorf("found = %v, want %v", found, tt.wantFound)
			}
			if !found {
				return
			}
			if got.CanCreateDerived != tt.wantEligible {
				t.Errorf("CanCreateDerived = %v, want %v", got.CanCreateDerived, tt.wantEligible)
			}
			if got.ReasonCode != tt.wantReason {
				t.Errorf("ReasonCode = %q, want %q", got.ReasonCode, tt.wantReason)
			}
			if tt.stage == "draft" && got.Conclusion != nil {
				t.Errorf("Conclusion = %+v, want nil for non-finalized state", got.Conclusion)
			}
			if tt.stage == "success" && (got.FixedConditions.Hypothesis != nil || len(got.FixedConditions.Prompts) != 2 || got.FixedConditions.Prompts[0].SequenceNo != 1) {
				t.Errorf("FixedConditions = %+v, want ordered prompts and nil hypothesis", got.FixedConditions)
			}
			if tt.stage == "hypothesis" && (got.FixedConditions.Hypothesis == nil || *got.FixedConditions.Hypothesis != "hypothesis") {
				t.Errorf("Hypothesis = %v, want hypothesis", got.FixedConditions.Hypothesis)
			}
		})
	}
}

var derivationSourceDriverOnce sync.Once

// driver注入store。
func newDerivationSourceStore(t *testing.T, stage string) *Store {
	t.Helper()
	derivationSourceDriverOnce.Do(func() {
		sql.Register("context-lab-derivation-source", derivationSourceDriver{})
	})
	db, err := sql.Open("context-lab-derivation-source", stage)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return &Store{db: db}
}

type derivationSourceDriver struct{}

func (derivationSourceDriver) Open(stage string) (driver.Conn, error) {
	return derivationSourceConn{stage: stage}, nil
}

type derivationSourceConn struct{ stage string }

func (derivationSourceConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unsupported")
}
func (derivationSourceConn) Close() error              { return nil }
func (derivationSourceConn) Begin() (driver.Tx, error) { return nil, errors.New("unsupported") }
func (c derivationSourceConn) QueryContext(_ context.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
	if strings.Contains(q, "FROM experiments e") {
		return c.main()
	}
	return c.prompts()
}
func (c derivationSourceConn) main() (driver.Rows, error) {
	if c.stage == "query-error" {
		return nil, errors.New("query")
	}
	r := &derivationRows{
		columns: []string{
			"id",
			"purpose",
			"id",
			"purpose",
			"hypothesis",
			"environment_conditions",
			"initial_input",
			"evaluation_axes",
			"fixed_at",
			"id",
			"conclusion",
			"state",
			"finalized_at",
		},
	}
	if c.stage == "missing" {
		return r, nil
	}
	v := []driver.Value{
		"experiment",
		"purpose",
		"fixed",
		"fixed purpose",
		nil,
		"env",
		"input",
		"axes",
		"2026-08-10T00:00:00Z",
		"conclusion",
		"text",
		"finalized",
		"2026-08-10T00:00:00Z",
	}
	if c.stage == "no-fixed" {
		for i := 2; i <= 8; i++ {
			v[i] = nil
		}
	}
	if c.stage == "no-conclusion" {
		for i := 9; i <= 12; i++ {
			v[i] = nil
		}
	}
	if c.stage == "draft" {
		v[11] = "draft"
	}
	if c.stage == "hypothesis" {
		v[4] = "hypothesis"
	}
	if c.stage == "fixed-time" {
		v[8] = "bad"
	}
	if c.stage == "conclusion-time" {
		v[12] = "bad"
	}
	if c.stage == "scan-error" {
		v[0] = nil
	}
	r.values = [][]driver.Value{v}
	return r, nil
}
func (c derivationSourceConn) prompts() (driver.Rows, error) {
	if c.stage == "prompt-query" {
		return nil, errors.New("prompt")
	}
	r := &derivationRows{
		columns: []string{
			"sequence_no",
			"content",
		},
	}
	if c.stage == "prompt-rows" {
		r.nextErr = errors.New("rows")
		return r, nil
	}
	if c.stage == "prompt-scan" {
		r.values = [][]driver.Value{
			{
				nil,
				"x",
			},
		}
		return r, nil
	}
	r.values = [][]driver.Value{
		{
			int64(1),
			"first",
		},
		{
			int64(2),
			"second",
		},
	}
	return r, nil
}

type derivationRows struct {
	columns []string
	values  [][]driver.Value
	index   int
	nextErr error
}

func (r *derivationRows) Columns() []string { return r.columns }
func (*derivationRows) Close() error        { return nil }
func (r *derivationRows) Next(d []driver.Value) error {
	if r.index >= len(r.values) {
		if r.nextErr != nil {
			return r.nextErr
		}
		return io.EOF
	}
	copy(d, r.values[r.index])
	r.index++
	return nil
}
