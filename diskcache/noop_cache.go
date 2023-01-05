package diskcache

import (
	"context"
)

type NoOp struct {
}

func NewNoOp() *NoOp {
	return &NoOp{}
}

func (n *NoOp) Store(ctx context.Context, key string, value []byte) error {
	return nil
}

func (n *NoOp) Load(ctx context.Context, key string) (value []byte, exist bool, err error) {
	return nil, false, nil
}

func (n *NoOp) Delete(ctx context.Context, key string) error {
	return nil
}
