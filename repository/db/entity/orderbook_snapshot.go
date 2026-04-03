package entity

type (
	OrderBookSnapshot struct {
		ID        int64            `db:"id"`
		Pair      string           `db:"pair"`
		Timestamp int64            `db:"timestamp"`
		Asks      []OrderBookEntry `db:"asks"`
		Bids      []OrderBookEntry `db:"bids"`
	}

	OrderBookEntry struct {
		Price  float64 `db:"price" json:"price"`
		Volume float64 `db:"volume" json:"volume"`
		Amount float64 `db:"amount" json:"amount"`
	}
)
