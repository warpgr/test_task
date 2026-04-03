package gateways

import (
	"context"
	"testing"
)

func TestGrinexAPIClient_GetOrderBook_Real(t *testing.T) {
	baseUrl := "https://grinex.io/api/v1"
	pair := "USDT"
	symbol := "usdta7a5"

	client := NewGrinexAPIClient(baseUrl, map[string]string{
		pair: symbol,
	})

	ob, err := client.GetOrderBook(context.Background(), pair)
	if err != nil {
		t.Fatalf("failed to get order book from real API: %v", err)
	}

	if ob == nil {
		t.Fatal("expected order book, got nil")
	}

	if ob.Timestamp == 0 {
		t.Error("expected non-zero timestamp")
	}

	// We expect at least some data from a live exchange
	if len(ob.Asks) == 0 && len(ob.Bids) == 0 {
		t.Error("expected at least one ask or bid in order book from live API")
	}

	t.Logf("Successfully fetched real order book for %s: %d asks, %d bids", pair, len(ob.Asks), len(ob.Bids))
}

func TestGrinexAPIClient_GetOrderBook_NotFound(t *testing.T) {
	client := NewGrinexAPIClient("https://grinex.io/api/v1", map[string]string{})
	_, err := client.GetOrderBook(context.Background(), "NONEXISTENT")
	if err == nil {
		t.Error("expected error for non-existent pair, got nil")
	}
}
