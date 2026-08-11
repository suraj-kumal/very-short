package contracts

import (
	"time"
)

type DirtyNode struct {
	Hash           string
	LastAccessTime time.Time
}
