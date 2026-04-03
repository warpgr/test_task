package gateways

import (
	"context"

	"github.com/warpgr/test_task/repository/db/entity"
)

type (
	ExchangeRatesAPIClient interface {
		GetOrderBook(ctx context.Context, pair string) (*entity.OrderBookSnapshot, error)
	}
)
