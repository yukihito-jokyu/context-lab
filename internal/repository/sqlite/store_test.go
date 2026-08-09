package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// SQLiteストア初期化検証
func TestOpen(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*testing.T) string
		wantError bool
		verify    func(*testing.T, *Store, string)
	}{
		{
			name: "管理ディレクトリとmigrationを初期化する",
			setup: func(t *testing.T) string {
				t.Helper()

				return filepath.Join(t.TempDir(), "context-lab")
			},
			verify: func(t *testing.T, store *Store, dataDirectory string) {
				t.Helper()

				if _, err := os.Stat(DatabasePath(dataDirectory)); err != nil {
					t.Errorf("database file stat error = %v", err)
				}
				var migrationCount int
				if err := store.db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&migrationCount); err != nil {
					t.Errorf("schema migrations query error = %v", err)
				}
				if migrationCount != 6 {
					t.Errorf("schema migrations count = %d, want %d", migrationCount, 6)
				}
				if err := store.migrate(context.Background()); err != nil {
					t.Errorf("migrate() error = %v", err)
				}
				if err := store.db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&migrationCount); err != nil {
					t.Errorf("schema migrations query after migrate error = %v", err)
				}
				if migrationCount != 6 {
					t.Errorf("schema migrations count after migrate = %d, want %d", migrationCount, 6)
				}
			},
		},
		{
			name: "管理ディレクトリがファイルなら失敗する",
			setup: func(t *testing.T) string {
				t.Helper()

				dataDirectory := filepath.Join(t.TempDir(), "not-a-directory")
				if err := os.WriteFile(dataDirectory, []byte("file"), 0o600); err != nil {
					t.Fatalf("write file error = %v", err)
				}

				return dataDirectory
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataDirectory := tt.setup(t)
			store, err := Open(dataDirectory)
			if gotError := err != nil; gotError != tt.wantError {
				t.Fatalf("Open() error = %v, want error = %v", err, tt.wantError)
			}
			if tt.wantError {
				if !strings.Contains(err.Error(), "create data directory") {
					t.Errorf("Open() error = %q, want create data directory error", err)
				}

				return
			}
			t.Cleanup(func() {
				if err := store.Close(); err != nil {
					t.Errorf("Close() error = %v", err)
				}
			})

			tt.verify(t, store, dataDirectory)
		})
	}
}

// SQLiteストア初期化失敗検証
func TestOpenFailures(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*testing.T)
		want    string
	}{
		{
			name: "データベースを開けない場合は失敗する",
			prepare: func(t *testing.T) {
				t.Helper()
				replaceOpenDatabase(t, func(string, string) (*sql.DB, error) {
					return nil, errors.New("open failed")
				})
			},
			want: "open database",
		},
		{
			name: "migrationを読み込めない場合は失敗する",
			prepare: func(t *testing.T) {
				t.Helper()
				replaceReadMigrationDirectory(t, func() ([]fs.DirEntry, error) {
					return nil, errors.New("read failed")
				})
			},
			want: "read migrations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare(t)

			_, err := Open(filepath.Join(t.TempDir(), "context-lab"))
			if err == nil {
				t.Fatal("Open() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Open() error = %q, want to contain %q", err, tt.want)
			}
		})
	}
}

// SQLiteストアclose検証
func TestStoreClose(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "SQLite接続を閉じる",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := Open(t.TempDir())
			if err != nil {
				t.Fatalf("Open() error = %v", err)
			}

			if err := store.Close(); err != nil {
				t.Errorf("Close() error = %v, want nil", err)
			}
		})
	}
}

// migration適用検証
func TestStoreApplyMigration(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	tests := []struct {
		name      string
		version   string
		wantError bool
	}{
		{
			name:    "適用済みmigrationは再実行しない",
			version: "000001_create_experiments.sql",
		},
		{
			name:      "存在しないmigrationは失敗する",
			version:   "missing.sql",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.applyMigration(context.Background(), tt.version)
			if gotError := err != nil; gotError != tt.wantError {
				t.Fatalf("applyMigration() error = %v, want error = %v", err, tt.wantError)
			}
		})
	}
}

// migration一覧読み出し検証
func TestStoreMigrate(t *testing.T) {
	tests := []struct {
		name    string
		entries []fs.DirEntry
		err     error
		stage   string
		wantErr bool
	}{
		{
			name:    "migration一覧の読み込み失敗を返す",
			err:     errors.New("read failed"),
			wantErr: true,
		},
		{
			name: "ディレクトリとSQL以外を除外する",
			entries: []fs.DirEntry{
				fakeMigrationEntry{
					name:      "directory",
					directory: true,
				},
				fakeMigrationEntry{name: "README.md"},
				fakeMigrationEntry{name: "000001_create_experiments.sql"},
			},
		},
		{
			name: "migration適用失敗を返す",
			entries: []fs.DirEntry{
				fakeMigrationEntry{name: "000001_create_experiments.sql"},
			},
			stage:   "create",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := Open(t.TempDir())
			if err != nil {
				t.Fatalf("Open() error = %v", err)
			}
			t.Cleanup(func() {
				if err := store.Close(); err != nil {
					t.Errorf("Close() error = %v", err)
				}
			})
			replaceReadMigrationDirectory(t, func() ([]fs.DirEntry, error) {
				return tt.entries, tt.err
			})
			if tt.stage != "" {
				if err := store.Close(); err != nil {
					t.Fatalf("Close() error = %v", err)
				}
				store.db = openMigrationFailureDatabase(t, tt.stage)
			}

			err = store.migrate(context.Background())
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Errorf("migrate() error = %v, want error = %v", err, tt.wantErr)
			}
		})
	}
}

// migrationのデータベース失敗検証
func TestStoreApplyMigrationDatabaseFailures(t *testing.T) {
	tests := []struct {
		name  string
		stage string
		want  string
	}{
		{
			name:  "migration記録テーブルの作成失敗",
			stage: "create",
			want:  "create migration table",
		},
		{
			name:  "適用済み確認の失敗",
			stage: "query",
			want:  "check migration",
		},
		{
			name:  "transaction開始の失敗",
			stage: "begin",
			want:  "begin migration",
		},
		{
			name:  "migration実行の失敗",
			stage: "apply",
			want:  "apply migration",
		},
		{
			name:  "migration記録の失敗",
			stage: "record",
			want:  "record migration",
		},
		{
			name:  "transaction確定の失敗",
			stage: "commit",
			want:  "commit migration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &Store{db: openMigrationFailureDatabase(t, tt.stage)}
			t.Cleanup(func() {
				if err := store.Close(); err != nil {
					t.Errorf("Close() error = %v", err)
				}
			})

			err := store.applyMigration(context.Background(), "000001_create_experiments.sql")
			if err == nil {
				t.Fatal("applyMigration() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("applyMigration() error = %q, want to contain %q", err, tt.want)
			}
		})
	}
}

// 実験一覧読み出し失敗検証
func TestStoreListByCancellationFailures(t *testing.T) {
	tests := []struct {
		name  string
		stage string
		want  string
	}{
		{
			name:  "query失敗を返す",
			stage: "list-query",
			want:  "query experiments",
		},
		{
			name:  "行走査失敗を返す",
			stage: "list-scan",
			want:  "scan experiment",
		},
		{
			name:  "行反復失敗を返す",
			stage: "list-iterate",
			want:  "iterate experiments",
		},
		{
			name:  "行close失敗を返す",
			stage: "list-close",
			want:  "close experiments rows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &Store{db: openMigrationFailureDatabase(t, tt.stage)}
			t.Cleanup(func() {
				if err := store.Close(); err != nil {
					t.Errorf("Close() error = %v", err)
				}
			})

			_, err := store.listByCancellation(context.Background(), false)
			if err == nil {
				t.Fatal("listByCancellation() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("listByCancellation() error = %q, want to contain %q", err, tt.want)
			}
		})
	}
}

// SQLiteデータベースパス検証
func TestDatabasePath(t *testing.T) {
	tests := []struct {
		name          string
		dataDirectory string
		want          string
	}{
		{
			name:          "管理ディレクトリ配下の固定名を返す",
			dataDirectory: "/tmp/context-lab",
			want:          "/tmp/context-lab/context-lab.sqlite",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DatabasePath(tt.dataDirectory); got != tt.want {
				t.Errorf("DatabasePath() = %q, want %q", got, tt.want)
			}
		})
	}
}

// データベースopen差し替え
func replaceOpenDatabase(t *testing.T, replacement func(string, string) (*sql.DB, error)) {
	t.Helper()

	original := openDatabase
	openDatabase = replacement
	t.Cleanup(func() {
		openDatabase = original
	})
}

// migration一覧読み出し差し替え
func replaceReadMigrationDirectory(t *testing.T, replacement func() ([]fs.DirEntry, error)) {
	t.Helper()

	original := readMigrationDirectory
	readMigrationDirectory = replacement
	t.Cleanup(func() {
		readMigrationDirectory = original
	})
}

// migration失敗再現用データベース生成
func openMigrationFailureDatabase(t *testing.T, stage string) *sql.DB {
	t.Helper()

	migrationFailureDriverOnce.Do(func() {
		sql.Register(migrationFailureDriverName, migrationFailureDriver{})
	})
	db, err := sql.Open(migrationFailureDriverName, stage)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}

	return db
}

const migrationFailureDriverName = "context-lab-migration-failure"

var migrationFailureDriverOnce sync.Once

// migration失敗再現用driver
type migrationFailureDriver struct{}

// Open は失敗段階を持つconnectionを返す。
func (migrationFailureDriver) Open(stage string) (driver.Conn, error) {
	return &migrationFailureConnection{stage: stage}, nil
}

// migration失敗再現用connection
type migrationFailureConnection struct {
	stage string
}

// Prepare は未使用のstatementを拒否する。
func (c *migrationFailureConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not supported")
}

// Close はconnectionを閉じる。
func (c *migrationFailureConnection) Close() error {
	return nil
}

// Begin は互換用transactionを開始する。
func (c *migrationFailureConnection) Begin() (driver.Tx, error) {
	return c.begin()
}

// BeginTx はcontext付きtransactionを開始する。
func (c *migrationFailureConnection) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return c.begin()
}

// ExecContext は指定段階の実行失敗を返す。
func (c *migrationFailureConnection) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	if c.stage == "create" && strings.HasPrefix(query, "CREATE TABLE IF NOT EXISTS schema_migrations") {
		return nil, errors.New("create failed")
	}
	if c.stage == "apply" && strings.Contains(query, "CREATE TABLE IF NOT EXISTS experiments") {
		return nil, errors.New("apply failed")
	}
	if c.stage == "record" && strings.HasPrefix(query, "INSERT INTO schema_migrations") {
		return nil, errors.New("record failed")
	}

	return driver.RowsAffected(1), nil
}

// QueryContext は適用済み確認の結果を返す。
func (c *migrationFailureConnection) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	if c.stage == "list-query" {
		return nil, errors.New("query failed")
	}
	if c.stage == "list-scan" {
		return &migrationFailureRows{
			columns: []string{"id"},
			values:  [][]driver.Value{{"experiment-1"}},
		}, nil
	}
	if c.stage == "list-iterate" {
		return &migrationFailureRows{
			columns: []string{"id"},
			nextErr: errors.New("iterate failed"),
		}, nil
	}
	if c.stage == "list-close" {
		return &migrationFailureRows{
			columns:  []string{"id"},
			keepOpen: true,
			closeErr: errors.New("close failed"),
		}, nil
	}
	if c.stage == "query" {
		return nil, errors.New("query failed")
	}

	return &migrationFailureRows{
		columns: []string{"exists"},
		values:  [][]driver.Value{{false}},
	}, nil
}

// begin は指定段階のtransaction開始失敗を返す。
func (c *migrationFailureConnection) begin() (driver.Tx, error) {
	if c.stage == "begin" {
		return nil, errors.New("begin failed")
	}

	return migrationFailureTransaction{stage: c.stage}, nil
}

// migration失敗再現用transaction
type migrationFailureTransaction struct {
	stage string
}

// Commit は指定段階の確定失敗を返す。
func (t migrationFailureTransaction) Commit() error {
	if t.stage == "commit" {
		return errors.New("commit failed")
	}

	return nil
}

// Rollback はtransactionを破棄する。
func (migrationFailureTransaction) Rollback() error {
	return nil
}

// migration適用済み確認用rows
type migrationFailureRows struct {
	columns  []string
	values   [][]driver.Value
	index    int
	nextErr  error
	keepOpen bool
	closeErr error
}

// Columns は適用済み真偽値の列を返す。
func (r *migrationFailureRows) Columns() []string {
	return r.columns
}

// Close はrowsを閉じる。
func (r *migrationFailureRows) Close() error {
	return r.closeErr
}

// HasNextResultSet は明示的なCloseまでrowsを開いたままにする。
func (r *migrationFailureRows) HasNextResultSet() bool {
	return r.keepOpen
}

// NextResultSet は追加の結果セットがないことを返す。
func (r *migrationFailureRows) NextResultSet() error {
	return io.EOF
}

// Next は未適用の真偽値を一度だけ返す。
func (r *migrationFailureRows) Next(destination []driver.Value) error {
	if r.nextErr != nil {
		return r.nextErr
	}
	if r.index == len(r.values) {
		return io.EOF
	}
	copy(destination, r.values[r.index])
	r.index++

	return nil
}

// migration一覧用entry
type fakeMigrationEntry struct {
	name      string
	directory bool
}

// Name はentry名を返す。
func (e fakeMigrationEntry) Name() string {
	return e.name
}

// IsDir はディレクトリかを返す。
func (e fakeMigrationEntry) IsDir() bool {
	return e.directory
}

// Type はentry種別を返す。
func (e fakeMigrationEntry) Type() fs.FileMode {
	if e.directory {
		return fs.ModeDir
	}

	return 0
}

// Info は本テストで使わない情報要求を拒否する。
func (e fakeMigrationEntry) Info() (fs.FileInfo, error) {
	return nil, errors.New("info is not supported")
}
