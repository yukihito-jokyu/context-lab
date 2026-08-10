package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"
)

// SQLite環境準備session一覧の種別分離と時刻順。
func TestStoreListPreparations(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	older := time.Date(2026, time.August, 10, 1, 0, 0, 0, time.UTC)
	newer := older.Add(time.Minute)
	sessions := []struct {
		id        string
		kind      string
		state     string
		createdAt time.Time
		updatedAt time.Time
	}{
		{
			id:        "preparation-old",
			kind:      environmentPreparationKind,
			state:     "completed",
			createdAt: older.Add(-time.Minute),
			updatedAt: older,
		},
		{
			id:        "briefing-session",
			kind:      "experiment_brief",
			state:     "started",
			createdAt: newer,
			updatedAt: newer.Add(time.Minute),
		},
		{
			id:        "preparation-new",
			kind:      environmentPreparationKind,
			state:     "running",
			createdAt: newer.Add(-time.Minute),
			updatedAt: newer,
		},
		{
			id:        "preparation-alpha",
			kind:      environmentPreparationKind,
			state:     "completed",
			createdAt: newer.Add(-time.Minute),
			updatedAt: newer,
		},
	}
	for _, session := range sessions {
		if _, err := store.db.Exec(
			"INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
			session.id,
			session.kind,
			session.state,
			session.createdAt.Format(time.RFC3339Nano),
			session.updatedAt.Format(time.RFC3339Nano),
		); err != nil {
			t.Fatalf("seed preparation session error = %v", err)
		}
	}

	got, err := store.ListPreparations(context.Background())
	if err != nil {
		t.Fatalf("ListPreparations() error = %v", err)
	}
	if gotCount := len(got); gotCount != 3 {
		t.Fatalf("preparations length = %d, want 3", gotCount)
	}
	if got[0].ID != "preparation-alpha" {
		t.Errorf("first Preparation.ID = %q, want %q", got[0].ID, "preparation-alpha")
	}
	if got[0].State != "completed" {
		t.Errorf("first Preparation.State = %q, want %q", got[0].State, "completed")
	}
	if !got[0].StartedAt.Equal(newer.Add(-time.Minute)) {
		t.Errorf("first Preparation.StartedAt = %s, want %s", got[0].StartedAt, newer.Add(-time.Minute))
	}
	if !got[0].LastObservedAt.Equal(newer) {
		t.Errorf("first Preparation.LastObservedAt = %s, want %s", got[0].LastObservedAt, newer)
	}
	if got[1].ID != "preparation-new" {
		t.Errorf("second Preparation.ID = %q, want %q", got[1].ID, "preparation-new")
	}
	if got[2].ID != "preparation-old" {
		t.Errorf("third Preparation.ID = %q, want %q", got[2].ID, "preparation-old")
	}
}

// SQLite環境準備session一覧の空配列返却。
func TestStoreListPreparationsReturnsEmptySlice(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	got, err := store.ListPreparations(context.Background())
	if err != nil {
		t.Fatalf("ListPreparations() error = %v", err)
	}
	if got == nil {
		t.Error("ListPreparations() = nil, want empty slice")
	}
	if gotCount := len(got); gotCount != 0 {
		t.Errorf("preparations length = %d, want 0", gotCount)
	}
}

// SQLite環境準備session詳細の種別分離と関連情報読み出し。
func TestStoreGetPreparation(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	now := time.Date(2026, time.August, 10, 1, 2, 3, 0, time.UTC)
	for _, query := range []struct {
		statement string
		arguments []any
	}{
		{
			statement: "INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
			arguments: []any{
				"preparation-1",
				environmentPreparationKind,
				"running",
				now.Format(time.RFC3339Nano),
				now.Add(time.Minute).Format(time.RFC3339Nano),
			},
		},
		{
			statement: "INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
			arguments: []any{
				"other-1",
				"experiment_brief",
				"running",
				now.Format(time.RFC3339Nano),
				now.Format(time.RFC3339Nano),
			},
		},
		{
			statement: "INSERT INTO environment_preparation_candidates (id, preparation_session_id, environment_conditions, safe_summary, created_at) VALUES (?, ?, ?, ?, ?)",
			arguments: []any{
				"candidate-1",
				"preparation-1",
				"macOS",
				"利用可能",
				now.Format(time.RFC3339Nano),
			},
		},
		{
			statement: "INSERT INTO environment_preparation_diagnostics (id, preparation_session_id, code, safe_summary, occurred_at) VALUES (?, ?, ?, ?, ?)",
			arguments: []any{
				"diagnostic-1",
				"preparation-1",
				"CHECKED",
				"確認済み",
				now.Format(time.RFC3339Nano),
			},
		},
		{
			statement: "INSERT INTO environment_preparation_operations (request_id, preparation_session_id, state, failure_code, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			arguments: []any{
				"request-1",
				"preparation-1",
				"failed",
				"PREPARATION_FAILED",
				now.Format(time.RFC3339Nano),
				now.Add(time.Minute).Format(time.RFC3339Nano),
			},
		},
	} {
		if _, err := store.db.Exec(query.statement, query.arguments...); err != nil {
			t.Fatalf("seed preparation detail error = %v", err)
		}
	}

	got, found, err := store.GetPreparation(context.Background(), "preparation-1")
	if err != nil {
		t.Fatalf("GetPreparation() error = %v", err)
	}
	if !found {
		t.Fatal("found = false, want true")
	}
	if got.Reconciliation.State != "running" {
		t.Errorf("Reconciliation.State = %q, want %q", got.Reconciliation.State, "running")
	}
	if len(got.Candidates) != 1 || got.Candidates[0].Summary != "利用可能" {
		t.Errorf("Candidates = %+v, want one safe candidate", got.Candidates)
	}
	if len(got.Diagnostics) != 1 || got.Diagnostics[0].SafeSummary != "確認済み" {
		t.Errorf("Diagnostics = %+v, want one safe diagnostic", got.Diagnostics)
	}
	if got.Failure == nil || got.Failure.Code != "PREPARATION_FAILED" {
		t.Errorf("Failure = %+v, want PREPARATION_FAILED", got.Failure)
	}

	_, found, err = store.GetPreparation(context.Background(), "other-1")
	if err != nil {
		t.Fatalf("GetPreparation() other kind error = %v", err)
	}
	if found {
		t.Error("other kind found = true, want false")
	}
}

// 環境準備詳細queryの全読み出し経路。
func TestGetPreparationPaths(t *testing.T) {
	validTime := "2026-08-10T01:02:03Z"
	tests := []struct {
		name      string
		queryer   stubPreparationQueryer
		wantFound bool
		wantError string
		wantState string
	}{
		{
			name:      "sessionが見つからない",
			queryer:   stubPreparationQueryer{session: stubPreparationRow{err: sql.ErrNoRows}},
			wantFound: false,
		},
		{
			name:      "session query失敗",
			queryer:   stubPreparationQueryer{session: stubPreparationRow{err: errors.New("session failed")}},
			wantError: "get preparation session",
		},
		{
			name: "開始時刻不正",
			queryer: stubPreparationQueryer{session: stubPreparationRow{values: []string{
				"preparation-1",
				"completed",
				"invalid",
				validTime,
			}}},
			wantError: "parse preparation start time",
		},
		{
			name: "最終観測時刻不正",
			queryer: stubPreparationQueryer{session: stubPreparationRow{values: []string{
				"preparation-1",
				"completed",
				validTime,
				"invalid",
			}}},
			wantError: "parse preparation observation time",
		},
		{
			name:      "候補query失敗",
			queryer:   successfulPreparationQueryer(validTime, "completed", stubPreparationRows{}, stubPreparationRows{}),
			wantError: "get preparation candidates",
		},
		{
			name:      "候補scan失敗",
			queryer:   successfulPreparationQueryer(validTime, "completed", stubPreparationRows{scanErr: errors.New("scan failed")}, stubPreparationRows{}),
			wantError: "scan preparation candidate",
		},
		{
			name: "候補時刻不正",
			queryer: successfulPreparationQueryer(validTime, "completed", stubPreparationRows{values: [][]string{
				{
					"candidate-1",
					"macOS",
					"safe",
					"invalid",
				},
			}}, stubPreparationRows{}),
			wantError: "parse preparation candidate time",
		},
		{
			name:      "候補反復失敗",
			queryer:   successfulPreparationQueryer(validTime, "completed", stubPreparationRows{err: errors.New("iterate failed")}, stubPreparationRows{}),
			wantError: "iterate preparation candidates",
		},
		{
			name:      "診断query失敗",
			queryer:   successfulPreparationQueryer(validTime, "completed", stubPreparationRows{}, stubPreparationRows{}),
			wantError: "get preparation diagnostics",
		},
		{
			name:      "診断scan失敗",
			queryer:   successfulPreparationQueryer(validTime, "completed", stubPreparationRows{}, stubPreparationRows{scanErr: errors.New("scan failed")}),
			wantError: "scan preparation diagnostic",
		},
		{
			name: "診断時刻不正",
			queryer: successfulPreparationQueryer(validTime, "completed", stubPreparationRows{}, stubPreparationRows{values: [][]string{
				{
					"diagnostic-1",
					"CHECKED",
					"safe",
					"invalid",
				},
			}}),
			wantError: "parse preparation diagnostic time",
		},
		{
			name:      "診断反復失敗",
			queryer:   successfulPreparationQueryer(validTime, "completed", stubPreparationRows{}, stubPreparationRows{err: errors.New("iterate failed")}),
			wantError: "iterate preparation diagnostics",
		},
		{
			name: "失敗query失敗",
			queryer: stubPreparationQueryer{
				session:     successfulPreparationSession(validTime, "completed"),
				candidates:  stubPreparationRows{},
				diagnostics: stubPreparationRows{},
				failure:     stubPreparationRow{err: errors.New("failure query failed")},
			},
			wantError: "get preparation operation",
		},
		{
			name: "失敗時刻不正",
			queryer: stubPreparationQueryer{
				session:     successfulPreparationSession(validTime, "completed"),
				candidates:  stubPreparationRows{},
				diagnostics: stubPreparationRows{},
				failure: stubPreparationRow{values: []string{
					"running",
					"",
					"invalid",
				}},
			},
			wantError: "parse preparation operation time",
		},
		{
			name: "完了状態を返す",
			queryer: stubPreparationQueryer{
				session:     successfulPreparationSession(validTime, "completed"),
				candidates:  stubPreparationRows{},
				diagnostics: stubPreparationRows{},
				failure:     stubPreparationRow{err: sql.ErrNoRows},
			},
			wantFound: true,
			wantState: "completed",
		},
		{
			name: "実行中を再照合状態へ変換する",
			queryer: stubPreparationQueryer{
				session:     successfulPreparationSession(validTime, "running"),
				candidates:  stubPreparationRows{},
				diagnostics: stubPreparationRows{},
				failure: stubPreparationRow{values: []string{
					"running",
					"",
					validTime,
				}},
			},
			wantFound: true,
			wantState: "reconciling",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "候補query失敗" {
				tt.queryer.candidatesErr = errors.New("candidate query failed")
			}
			if tt.name == "診断query失敗" {
				tt.queryer.diagnosticsErr = errors.New("diagnostic query failed")
			}
			got, found, err := getPreparation(context.Background(), tt.queryer, "preparation-1")
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("getPreparation() error = %v, want containing %q", err, tt.wantError)
				}

				return
			}
			if err != nil {
				t.Fatalf("getPreparation() error = %v", err)
			}
			if found != tt.wantFound {
				t.Errorf("found = %v, want %v", found, tt.wantFound)
			}
			if tt.wantState != "" && got.Reconciliation.State != tt.wantState {
				t.Errorf("Reconciliation.State = %q, want %q", got.Reconciliation.State, tt.wantState)
			}
		})
	}
}

// SQLite環境準備詳細の最新操作状態反映。
func TestStoreGetPreparationUsesLatestOperation(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	now := time.Date(2026, time.August, 10, 1, 2, 3, 0, time.UTC)
	if _, err := store.db.Exec("INSERT INTO preparation_sessions (id, kind, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", "preparation-1", environmentPreparationKind, "completed", now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)); err != nil {
		t.Fatalf("seed session error = %v", err)
	}
	if _, err := store.db.Exec("INSERT INTO environment_preparation_operations (request_id, preparation_session_id, state, failure_code, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", "request-old", "preparation-1", "failed", "PREPARATION_FAILED", now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)); err != nil {
		t.Fatalf("seed failed operation error = %v", err)
	}
	if _, err := store.db.Exec("INSERT INTO environment_preparation_operations (request_id, preparation_session_id, state, failure_code, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", "request-new", "preparation-1", "running", nil, now.Add(time.Minute).Format(time.RFC3339Nano), now.Add(time.Minute).Format(time.RFC3339Nano)); err != nil {
		t.Fatalf("seed running operation error = %v", err)
	}

	got, found, err := store.GetPreparation(context.Background(), "preparation-1")
	if err != nil {
		t.Fatalf("GetPreparation() error = %v", err)
	}
	if !found {
		t.Fatal("found = false, want true")
	}
	if got.Reconciliation.State != "reconciling" {
		t.Errorf("Reconciliation.State = %q, want %q", got.Reconciliation.State, "reconciling")
	}
	if got.Failure != nil {
		t.Errorf("Failure = %+v, want nil after latest running operation", got.Failure)
	}
}

// stubPreparationQueryer は環境準備詳細queryのtest double。
type stubPreparationQueryer struct {
	session        stubPreparationRow
	candidates     stubPreparationRows
	candidatesErr  error
	diagnostics    stubPreparationRows
	diagnosticsErr error
	failure        stubPreparationRow
}

// QueryContext はquery種別ごとの指定済み行または失敗を返す。
func (q stubPreparationQueryer) QueryContext(_ context.Context, query string, _ ...any) (preparationRows, error) {
	if strings.Contains(query, "environment_preparation_candidates") {
		return &q.candidates, q.candidatesErr
	}

	return &q.diagnostics, q.diagnosticsErr
}

// QueryRowContext はquery種別ごとの指定済み行を返す。
func (q stubPreparationQueryer) QueryRowContext(_ context.Context, query string, _ ...any) preparationRow {
	if strings.Contains(query, "preparation_sessions") {
		return q.session
	}

	return q.failure
}

// stubPreparationRow は環境準備詳細の単一行test double。
type stubPreparationRow struct {
	values []string
	err    error
}

// Scan は指定済み値または失敗を返す。
func (r stubPreparationRow) Scan(destinations ...any) error {
	if r.err != nil {
		return r.err
	}
	for index, destination := range destinations {
		pointer, ok := destination.(*string)
		if !ok {
			return errors.New("unexpected scan destination")
		}
		*pointer = r.values[index]
	}

	return nil
}

// successfulPreparationQueryer は成功するsessionと関連行test doubleを返す。
func successfulPreparationQueryer(timestamp, state string, candidates, diagnostics stubPreparationRows) stubPreparationQueryer {
	return stubPreparationQueryer{
		session:     successfulPreparationSession(timestamp, state),
		candidates:  candidates,
		diagnostics: diagnostics,
		failure:     stubPreparationRow{err: sql.ErrNoRows},
	}
}

// successfulPreparationSession は成功する環境準備session行を返す。
func successfulPreparationSession(timestamp, state string) stubPreparationRow {
	return stubPreparationRow{values: []string{
		"preparation-1",
		state,
		timestamp,
		timestamp,
	}}
}

// SQLite環境準備session一覧の読み出し失敗。
func TestStoreListPreparationsFailures(t *testing.T) {
	tests := []struct {
		name string
		rows *stubPreparationRows
		err  error
		want string
	}{
		{
			name: "query失敗を返す",
			err:  errors.New("query failed"),
			want: "list preparations",
		},
		{
			name: "scan失敗を返す",
			rows: &stubPreparationRows{scanErr: errors.New("scan failed")},
			want: "scan preparation",
		},
		{
			name: "開始時刻不正を返す",
			rows: &stubPreparationRows{values: [][]string{
				{
					"preparation-1",
					"running",
					"invalid",
					"2026-08-10T00:00:00Z",
				},
			}},
			want: "parse preparation start time",
		},
		{
			name: "最終観測時刻不正を返す",
			rows: &stubPreparationRows{values: [][]string{
				{
					"preparation-1",
					"running",
					"2026-08-10T00:00:00Z",
					"invalid",
				},
			}},
			want: "parse preparation observation time",
		},
		{
			name: "行反復失敗を返す",
			rows: &stubPreparationRows{err: errors.New("iterate failed")},
			want: "iterate preparations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &Store{listPreparations: func(context.Context) (preparationRows, error) {
				return tt.rows, tt.err
			}}

			_, err := store.ListPreparations(context.Background())
			if err == nil {
				t.Fatal("ListPreparations() error = nil, want error")
			}
			if got := err.Error(); !strings.Contains(got, tt.want) {
				t.Errorf("ListPreparations() error = %q, want containing %q", got, tt.want)
			}
		})
	}
}

// stubPreparationRows は環境準備一覧repositoryのtest double。
type stubPreparationRows struct {
	values  [][]string
	index   int
	scanErr error
	err     error
}

// Close は行読み出し終了を返す。
func (s *stubPreparationRows) Close() error {
	return nil
}

// Err は指定済みの反復エラーを返す。
func (s *stubPreparationRows) Err() error {
	return s.err
}

// Next は次の指定済み行の有無を返す。
func (s *stubPreparationRows) Next() bool {
	if s.scanErr != nil {
		return s.index == 0
	}
	if s.index >= len(s.values) {
		return false
	}

	return true
}

// Scan は現在行の文字列を読み出す。
func (s *stubPreparationRows) Scan(destinations ...any) error {
	if s.scanErr != nil {
		s.index++

		return s.scanErr
	}
	values := s.values[s.index]
	s.index++
	for index, destination := range destinations {
		pointer, ok := destination.(*string)
		if !ok {
			return errors.New("unexpected scan destination")
		}
		*pointer = values[index]
	}

	return nil
}
