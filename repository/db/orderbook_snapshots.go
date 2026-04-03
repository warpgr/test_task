package db

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/warpgr/test_task/repository/db/entity"
)

type orderBookSnapshotsDB struct {
	conn *sqlx.DB
}

func NewOrderBookSnapshotsDB(conn *sqlx.DB) OrderBookSnapshotsDB {
	return &orderBookSnapshotsDB{conn: conn}
}

func (db *orderBookSnapshotsDB) GetOrderBookSnapshot(ctx context.Context, pair string) (*entity.OrderBookSnapshot, error) {
	const query = `SELECT id, pair, timestamp_ms, payload FROM orderbook_snapshots WHERE pair = $1 ORDER BY timestamp_ms DESC LIMIT 1`

	var row struct {
		ID        int64  `db:"id"`
		Pair      string `db:"pair"`
		Timestamp int64  `db:"timestamp_ms"`
		Payload   []byte `db:"payload"`
	}

	if err := db.conn.GetContext(ctx, &row, query, pair); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	type payload struct {
		Asks []entity.OrderBookEntry `json:"asks"`
		Bids []entity.OrderBookEntry `json:"bids"`
	}

	var p payload
	if err := json.Unmarshal(row.Payload, &p); err != nil {
		return nil, err
	}

	return &entity.OrderBookSnapshot{
		ID:        row.ID,
		Pair:      row.Pair,
		Timestamp: row.Timestamp,
		Asks:      p.Asks,
		Bids:      p.Bids,
	}, nil
}

func (db *orderBookSnapshotsDB) SaveOrderBookSnapshot(ctx context.Context, snapshot *entity.OrderBookSnapshot) error {
	const query = `INSERT INTO orderbook_snapshots (pair, timestamp_ms, payload) VALUES ($1, $2, $3)`
	type payload struct {
		Asks []entity.OrderBookEntry `json:"asks"`
		Bids []entity.OrderBookEntry `json:"bids"`
	}

	p := payload{
		Asks: snapshot.Asks,
		Bids: snapshot.Bids,
	}

	pJSON, err := json.Marshal(p)
	if err != nil {
		return err
	}

	_, err = db.conn.ExecContext(ctx, query, snapshot.Pair, snapshot.Timestamp, pJSON)
	return err
}

func (db *orderBookSnapshotsDB) SaveOrderBookSnapshots(ctx context.Context, snapshots []*entity.OrderBookSnapshot) error {
	const query = `INSERT INTO orderbook_snapshots (pair, timestamp_ms, payload) VALUES ($1, $2, $3)`
	type payload struct {
		Asks []entity.OrderBookEntry `json:"asks"`
		Bids []entity.OrderBookEntry `json:"bids"`
	}

	tx, err := db.conn.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, snapshot := range snapshots {
		p := payload{
			Asks: snapshot.Asks,
			Bids: snapshot.Bids,
		}

		pJSON, err := json.Marshal(p)
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, query, snapshot.Pair, snapshot.Timestamp, pJSON); err != nil {
			return err
		}
	}

	return tx.Commit()
}
