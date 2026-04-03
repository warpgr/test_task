package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/warpgr/test_task/repository/db/entity"
)

type OrdersbookSnapshotMockDB struct {
	mock.Mock
}

func (m *OrdersbookSnapshotMockDB) GetOrderBookSnapshot(ctx context.Context, pair string) (*entity.OrderBookSnapshot, error) {
	args := m.Called(ctx, pair)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.OrderBookSnapshot), args.Error(1)
}
