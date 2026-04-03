package gateways

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/warpgr/test_task/repository/db/entity"
)

func NewGrinexAPIClient(baseUrl string, pairsUrls map[string]string) ExchangeRatesAPIClient {
	return &grinexAPIClient{
		baseUrl:   baseUrl,
		pairsUrls: pairsUrls,
		client:    resty.New(),
	}
}

type (
	grinexAPIClient struct {
		baseUrl   string
		pairsUrls map[string]string
		client    *resty.Client
	}

	grinexOrderBookResponse struct {
		Timestamp int64        `json:"timestamp"`
		Asks      []grinexItem `json:"asks"`
		Bids      []grinexItem `json:"bids"`
	}

	grinexItem struct {
		Price  string `json:"price"`
		Volume string `json:"volume"`
		Amount string `json:"amount"`
	}
)

func (c *grinexAPIClient) GetOrderBook(ctx context.Context, pair string) (*entity.OrderBookSnapshot, error) {
	const (
		spotDepthEndpoint = "/spot/depth"
	)

	symbol, ok := c.pairsUrls[pair]
	if !ok {
		return nil, fmt.Errorf("pair %s not found", pair)
	}

	fullUrl, err := url.JoinPath(c.baseUrl, spotDepthEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to join url: %w", err)
	}

	var grinexResp grinexOrderBookResponse
	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("symbol", symbol).
		SetResult(&grinexResp).
		Get(fullUrl)

	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	orderBook := &entity.OrderBookSnapshot{
		Pair:      pair,
		Timestamp: grinexResp.Timestamp,
		Asks:      make([]entity.OrderBookEntry, 0, len(grinexResp.Asks)),
		Bids:      make([]entity.OrderBookEntry, 0, len(grinexResp.Bids)),
	}

	for _, item := range grinexResp.Asks {
		entry, err := parseGrinexItem(item)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ask item: %w", err)
		}
		orderBook.Asks = append(orderBook.Asks, entry)
	}

	for _, item := range grinexResp.Bids {
		entry, err := parseGrinexItem(item)
		if err != nil {
			return nil, fmt.Errorf("failed to parse bid item: %w", err)
		}
		orderBook.Bids = append(orderBook.Bids, entry)
	}

	return orderBook, nil
}

func parseGrinexItem(item grinexItem) (entity.OrderBookEntry, error) {
	price, err := strconv.ParseFloat(item.Price, 64)
	if err != nil {
		return entity.OrderBookEntry{}, fmt.Errorf("failed to parse price: %w", err)
	}
	volume, err := strconv.ParseFloat(item.Volume, 64)
	if err != nil {
		return entity.OrderBookEntry{}, fmt.Errorf("failed to parse volume: %w", err)
	}
	amount, err := strconv.ParseFloat(item.Amount, 64)
	if err != nil {
		return entity.OrderBookEntry{}, fmt.Errorf("failed to parse amount: %w", err)
	}

	return entity.OrderBookEntry{
		Price:  price,
		Volume: volume,
		Amount: amount,
	}, nil
}
