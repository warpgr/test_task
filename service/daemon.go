package service

import (
	"context"
	"time"

	"github.com/warpgr/test_task/repository/db"
	"github.com/warpgr/test_task/repository/gateways"
	"go.uber.org/zap"
)

func NewDaemon(orderBookSnapshotsDB db.OrderBookSnapshotsDBWriter, client gateways.ExchangeRatesAPIClient, pairs map[string]string, logger *zap.Logger) Daemon {
	return &daemon{
		orderBookSnapshotsDB: orderBookSnapshotsDB,
		client:               client,
		pairs:                pairs,
		logger:               logger,
	}
}

type daemon struct {
	orderBookSnapshotsDB db.OrderBookSnapshotsDBWriter
	client               gateways.ExchangeRatesAPIClient
	pairs                map[string]string
	logger               *zap.Logger
}

func (d *daemon) Run(ctx context.Context) <-chan error {
	errChan := make(chan error, 1)
	go func() {
		fetchPer := time.NewTicker(1 * time.Second)
		defer fetchPer.Stop()
		defer close(errChan)

		for {
			select {
			case <-ctx.Done():
				return

			case <-fetchPer.C:
				for pair := range d.pairs {
					orderBook, err := d.client.GetOrderBook(ctx, pair)
					if err != nil {
						d.logger.Error("failed to fetch order book", zap.String("pair", pair), zap.Error(err))
						errChan <- err
						return
					}
					if err := d.orderBookSnapshotsDB.SaveOrderBookSnapshot(ctx, orderBook); err != nil {
						d.logger.Error("failed to save order book snapshot", zap.String("pair", pair), zap.Error(err))
						errChan <- err
						return
					}
					d.logger.Info("successfully saved order book snapshot", zap.String("pair", pair))
				}
			}
		}

	}()
	return errChan
}
