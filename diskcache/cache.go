package diskcache

import (
	"context"
)

type Interface interface {
	Store(ctx context.Context, key string, value []byte) error
	Load(ctx context.Context, key string) (value []byte, exist bool, err error)
	Delete(ctx context.Context, key string) error
}
