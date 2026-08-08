package sqlite

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
)

// SQLite初期化と実験一覧読み出し。
func TestStoreListExperiments(t *testing.T) {
	tests := []struct {
		name                string
		seed                func(*testing.T, *Store)
		wantExperiments     []string
		wantCancelled       []string
		wantLastConfirmedAt bool
		wantDriverAvailable bool
	}{
		{
			name: "空のデータベースは空配列を返す",
			seed: func(t *testing.T, store *Store) {
				t.Helper()
			},
			wantExperiments:     []string{},
			wantCancelled:       []string{},
			wantLastConfirmedAt: true,
			wantDriverAvailable: true,
		},
		{
			name: "取消済みを別配列へ分離する",
			seed: func(t *testing.T, store *Store) {
				t.Helper()
				seedExperiments(t, store)
			},
			wantExperiments: []string{
				"experiment-running",
				"experiment-planned",
			},
			wantCancelled: []string{
				"experiment-cancelled",
			},
			wantLastConfirmedAt: true,
			wantDriverAvailable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore(t)
			tt.seed(t, store)

			var version string
			if err := store.db.QueryRow("SELECT sqlite_version()").Scan(&version); err != nil {
				t.Fatalf("sqlite_version() error = %v", err)
			}
			if gotDriverAvailable := version != ""; gotDriverAvailable != tt.wantDriverAvailable {
				t.Errorf("driver available = %v, want %v", gotDriverAvailable, tt.wantDriverAvailable)
			}

			got, err := store.ListExperiments(context.Background())
			if err != nil {
				t.Fatalf("ListExperiments() error = %v", err)
			}
			if gotIDs := experimentIDs(got.Experiments); !reflect.DeepEqual(gotIDs, tt.wantExperiments) {
				t.Errorf("Experiments IDs = %v, want %v", gotIDs, tt.wantExperiments)
			}
			if gotIDs := experimentIDs(got.CancelledExperiments); !reflect.DeepEqual(gotIDs, tt.wantCancelled) {
				t.Errorf("CancelledExperiments IDs = %v, want %v", gotIDs, tt.wantCancelled)
			}
			if gotLastConfirmedAt := got.LastConfirmedAt != nil; gotLastConfirmedAt != tt.wantLastConfirmedAt {
				t.Errorf("LastConfirmedAt available = %v, want %v", gotLastConfirmedAt, tt.wantLastConfirmedAt)
			}
			if got.LastConfirmedAt != nil {
				var persisted string
				if err := store.db.QueryRow("SELECT value FROM application_metadata WHERE key = ?", "last_confirmed_at").Scan(&persisted); err != nil {
					t.Fatalf("last_confirmed_at query error = %v", err)
				}
				if got := got.LastConfirmedAt.Format(time.RFC3339Nano); got != persisted {
					t.Errorf("LastConfirmedAt = %q, want persisted %q", got, persisted)
				}
			}
		})
	}
}

// SQLite読み出し失敗の安全なrepositoryエラー化。
func TestStoreListExperimentsErrors(t *testing.T) {
	tests := []struct {
		name string
		seed func(*testing.T, *Store)
	}{
		{
			name: "通常実験のquery失敗",
			seed: func(t *testing.T, store *Store) {
				t.Helper()
				if _, err := store.db.Exec("DROP TABLE experiments"); err != nil {
					t.Fatalf("DROP TABLE error = %v", err)
				}
			},
		},
		{
			name: "取消済み実験の変換失敗",
			seed: func(t *testing.T, store *Store) {
				t.Helper()
				if _, err := store.db.Exec("INSERT INTO experiments (id, purpose, state, progress_summary, updated_at) VALUES (?, ?, ?, ?, ?)", "cancelled", "中止", "cancelled", "中止", "invalid-time"); err != nil {
					t.Fatalf("INSERT cancelled experiment error = %v", err)
				}
			},
		},
		{
			name: "最終確認時刻の記録失敗",
			seed: func(t *testing.T, store *Store) {
				t.Helper()
				if _, err := store.db.Exec("DROP TABLE application_metadata"); err != nil {
					t.Fatalf("DROP TABLE metadata error = %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore(t)
			tt.seed(t, store)

			_, err := store.ListExperiments(context.Background())
			if err == nil {
				t.Error("ListExperiments() error = nil, want repository error")
			}
		})
	}
}

// 一時SQLiteストア生成。
func newTestStore(t *testing.T) *Store {
	t.Helper()

	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	return store
}

// 実験fixture投入。
func seedExperiments(t *testing.T, store *Store) {
	t.Helper()

	statements := []struct {
		query string
		args  []any
	}{
		{
			query: "INSERT INTO experiments (id, purpose, state, progress_summary, derived_from_experiment_id, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			args: []any{
				"experiment-planned",
				"計画",
				"planned",
				"未開始",
				nil,
				"2026-08-08T01:00:00Z",
			},
		},
		{
			query: "INSERT INTO experiments (id, purpose, state, progress_summary, derived_from_experiment_id, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			args: []any{
				"experiment-running",
				"実行中",
				"running",
				"評価中",
				"experiment-planned",
				"2026-08-08T02:00:00Z",
			},
		},
		{
			query: "INSERT INTO experiments (id, purpose, state, progress_summary, derived_from_experiment_id, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			args: []any{
				"experiment-cancelled",
				"中止済み",
				"cancelled",
				"中止",
				nil,
				"2026-08-08T03:00:00Z",
			},
		},
		{
			query: "INSERT INTO application_metadata (key, value) VALUES (?, ?)",
			args: []any{
				"last_confirmed_at",
				"2026-08-08T01:02:03Z",
			},
		},
	}

	for _, statement := range statements {
		if _, err := store.db.Exec(statement.query, statement.args...); err != nil {
			t.Fatalf("seed database error = %v", err)
		}
	}
}

// 実験ID抽出。
func experimentIDs(experiments []domain.Experiment) []string {
	ids := make([]string, 0, len(experiments))
	for _, experiment := range experiments {
		ids = append(ids, experiment.ID)
	}

	return ids
}
