package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	pb "github.com/warpgr/test_task/proto"
	"github.com/warpgr/test_task/repository/db/entity"
	"github.com/warpgr/test_task/repository/db/mocks"
	"github.com/warpgr/test_task/service"
	"go.uber.org/zap"
)

func TestGetRates_TopN(t *testing.T) {
	mockDB := &mocks.OrdersbookSnapshotMockDB{}
	svc := service.NewRatesService(mockDB, zap.NewNop())

	ctx := t.Context()
	pair := "BTC_USDT"

	snapshot := &entity.OrderBookSnapshot{
		Pair:      pair,
		Timestamp: 1000,
		Asks: []entity.OrderBookEntry{
			{Price: 100.0},
			{Price: 101.0},
		},
		Bids: []entity.OrderBookEntry{
			{Price: 99.0},
			{Price: 98.0},
		},
	}

	mockDB.On("GetOrderBookSnapshot", mock.Anything, pair).Return(snapshot, nil)

	req := &pb.GetRatesRequest{
		Pair:       pair,
		CalcMethod: pb.CalcMethod_CALC_METHOD_TOP_N,
		N:          1,
	}

	res, err := svc.GetRates(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, 101.0, res.Ask)
	assert.Equal(t, 98.0, res.Bid)
}

func TestGetRates_AvgNM(t *testing.T) {
	mockDB := &mocks.OrdersbookSnapshotMockDB{}
	svc := service.NewRatesService(mockDB, zap.NewNop())

	ctx := t.Context()
	pair := "BTC_USDT"

	snapshot := &entity.OrderBookSnapshot{
		Pair:      pair,
		Timestamp: 1000,
		Asks: []entity.OrderBookEntry{
			{Price: 100.0},
			{Price: 101.0},
			{Price: 102.0},
		},
		Bids: []entity.OrderBookEntry{
			{Price: 90.0},
			{Price: 91.0},
			{Price: 92.0},
		},
	}

	mockDB.On("GetOrderBookSnapshot", mock.Anything, pair).Return(snapshot, nil)

	req := &pb.GetRatesRequest{
		Pair:       pair,
		CalcMethod: pb.CalcMethod_CALC_METHOD_AVG_N_M,
		N:          0,
		M:          2,
	}

	res, err := svc.GetRates(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, 101.0, res.Ask)
	assert.Equal(t, 91.0, res.Bid)
}
