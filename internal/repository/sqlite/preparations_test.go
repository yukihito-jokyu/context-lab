package sqlite

import (
	"context"
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
