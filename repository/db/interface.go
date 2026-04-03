package db

import (
	"context"

	"github.com/warpgr/test_task/repository/db/entity"
)

type (
	OrderBookSnapshotsDBReader interface {
		GetOrderBookSnapshot(ctx context.Context, pair string) (*entity.OrderBookSnapshot, error)
	}

	OrderBookSnapshotsDBWriter interface {
		SaveOrderBookSnapshot(ctx context.Context, snapshot *entity.OrderBookSnapshot) error
		SaveOrderBookSnapshots(ctx context.Context, snapshots []*entity.OrderBookSnapshot) error
	}

	OrderBookSnapshotsDB interface {
		OrderBookSnapshotsDBReader
		OrderBookSnapshotsDBWriter
	}
)
