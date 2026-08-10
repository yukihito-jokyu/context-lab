package domain

import "time"

// Preparation は環境準備sessionの安全な一覧表示情報。
type Preparation struct {
	ID             string
	State          string
	StartedAt      time.Time
	LastObservedAt time.Time
}
