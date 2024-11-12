package diskcache

import (
	"context"
)

type NoOp struct {
}

func NewNoOp() *NoOp {
	return &NoOp{}
}

func (n *NoOp) Store(_ context.Context, _ string, _ []byte) error {
	return nil
}

func (n *NoOp) Load(_ context.Context, _ string) (value []byte, exist bool, err error) {
	return nil, false, nil
}

func (n *NoOp) Delete(_ context.Context, _ string) error {
	return nil
}
