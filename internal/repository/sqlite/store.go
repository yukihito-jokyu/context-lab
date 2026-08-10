package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/yukihito-jokyu/context-lab/internal/domain"
	"github.com/yukihito-jokyu/context-lab/migrations"
	_ "modernc.org/sqlite"
)

const (
	databaseFileName  = "context-lab.sqlite"
	sqliteBusyTimeout = 10 * time.Second
)

var createDataDirectory = os.MkdirAll

var openDatabase = sql.Open

var readMigrationDirectory = func() ([]fs.DirEntry, error) {
	return fs.ReadDir(migrations.Files, ".")
}

// Store はSQLiteの実験読み出しadapter。
type Store struct {
	db                           *sql.DB
	beginBriefingTransaction     func(context.Context) (briefingTransaction, error)
	failBriefingMessageOperation func(context.Context, string, string) (sql.Result, error)
	listPreparations             func(context.Context) (preparationRows, error)
	getPreparation               func(context.Context, string) (domain.PreparationDetail, bool, error)
	briefingMessageMu            sync.Mutex
	derivationBriefingMessageMu  sync.Mutex
	listMu                       sync.Mutex
}

// Open は管理ディレクトリとSQLiteスキーマを初期化。
func Open(dataDirectory string) (*Store, error) {
	if err := createDataDirectory(dataDirectory, 0o700); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	db, err := openDatabase("sqlite", databaseDSN(dataDirectory))
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	store := &Store{
		db: db,
		beginBriefingTransaction: func(ctx context.Context) (briefingTransaction, error) {
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				return nil, err
			}

			return sqliteBriefingTransaction{tx: tx}, nil
		},
		failBriefingMessageOperation: func(ctx context.Context, requestID, failureCode string) (sql.Result, error) {
			return db.ExecContext(ctx, "UPDATE briefing_message_operations SET state = ?, failure_code = ? WHERE request_id = ?", domain.BriefingStartStateFailed, failureCode, requestID)
		},
		listPreparations: func(ctx context.Context) (preparationRows, error) {
			rows, err := db.QueryContext(ctx, listPreparationsQuery, environmentPreparationKind)
			if err != nil {
				return nil, err
			}

			return sqlitePreparationRows{rows: rows}, nil
		},
		getPreparation: func(ctx context.Context, preparationID string) (domain.PreparationDetail, bool, error) {
			return getPreparation(ctx, sqlitePreparationQueryer{db: db}, preparationID)
		},
	}
	if err := store.migrate(context.Background()); err != nil {
		_ = db.Close()

		return nil, err
	}

	return store, nil
}

// databaseDSN は同時読込・書込を待機できるWAL接続文字列を返す。
func databaseDSN(dataDirectory string) string {
	query := url.Values{}
	query.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", sqliteBusyTimeout.Milliseconds()))
	query.Add("_pragma", "journal_mode(WAL)")

	return (&url.URL{
		Scheme:   "file",
		Path:     filepath.Join(dataDirectory, databaseFileName),
		RawQuery: query.Encode(),
	}).String()
}

// Close はSQLite接続を閉じる。
func (s *Store) Close() error {
	return s.db.Close()
}

// migrate は未適用migrationだけを順番に適用。
func (s *Store) migrate(ctx context.Context) error {
	entries, err := readMigrationDirectory()
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		if err := s.applyMigration(ctx, entry.Name()); err != nil {
			return err
		}
	}

	return nil
}

// applyMigration は一つの未適用migrationを記録付きで実行。
func (s *Store) applyMigration(ctx context.Context, version string) error {
	contents, err := migrations.Files.ReadFile(version)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", version, err)
	}

	if _, err := s.db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TEXT NOT NULL)"); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}

	var applied bool
	if err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)", version).Scan(&applied); err != nil {
		return fmt.Errorf("check migration %s: %w", version, err)
	}
	if applied {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", version, err)
	}

	if _, err := tx.ExecContext(ctx, string(contents)); err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("apply migration %s: %w", version, err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)", version, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("record migration %s: %w", version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", version, err)
	}

	return nil
}

// DatabasePath は管理ディレクトリ配下のSQLiteパスを返す。
func DatabasePath(dataDirectory string) string {
	return filepath.Join(dataDirectory, databaseFileName)
}
