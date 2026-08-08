package migrations

import "embed"

// Files は組み込み済みの前進専用マイグレーション群。
//
//go:embed *.sql
var Files embed.FS
