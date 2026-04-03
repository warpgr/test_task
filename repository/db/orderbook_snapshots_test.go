package db_test

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/warpgr/test_task/repository/db"
	"github.com/warpgr/test_task/repository/db/entity"
)

type OrderBookSnapshotsDBTestSuite struct {
	suite.Suite

	migrationsPath string

	dsn       string
	container *postgres.PostgresContainer
	conn      *sqlx.DB
}

func TestRunOrderBookSnapshotsDBTestSuite(t *testing.T) {
	migrationsPath, provided := os.LookupEnv("DB_MIGRATIONS_PATH")
	if !provided {
		t.Skip("Skipping OrderBookSnapshotsDBTestSuite DB_MIGRATIONS_PATH didn't provided.")
		return
	}
	suite.Run(t, &OrderBookSnapshotsDBTestSuite{
		migrationsPath: migrationsPath,
	})
}

func (s *OrderBookSnapshotsDBTestSuite) TestFlow() {
	repo := db.NewOrderBookSnapshotsDB(s.conn)
	ctx := s.T().Context()

	snapshot := &entity.OrderBookSnapshot{
		Pair:      "BTC_USDT",
		Timestamp: 1625097600000,
		Asks: []entity.OrderBookEntry{
			{Price: 35000.5, Volume: 0.1, Amount: 3500.05},
		},
		Bids: []entity.OrderBookEntry{
			{Price: 34999.5, Volume: 0.2, Amount: 6999.9},
		},
	}

	err := repo.SaveOrderBookSnapshot(ctx, snapshot)
	s.NoError(err)

	retrieved, err := repo.GetOrderBookSnapshot(ctx, "BTC_USDT")
	s.NoError(err)
	s.NotNil(retrieved)

	s.Equal(snapshot.Pair, retrieved.Pair)
	s.Equal(snapshot.Timestamp, retrieved.Timestamp)
	s.Equal(len(snapshot.Asks), len(retrieved.Asks))
	s.Equal(snapshot.Asks[0].Price, retrieved.Asks[0].Price)
	s.Equal(snapshot.Asks[0].Volume, retrieved.Asks[0].Volume)
	s.Equal(snapshot.Asks[0].Amount, retrieved.Asks[0].Amount)
	s.Equal(len(snapshot.Bids), len(retrieved.Bids))
	s.Equal(snapshot.Bids[0].Price, retrieved.Bids[0].Price)
	s.Equal(snapshot.Bids[0].Volume, retrieved.Bids[0].Volume)
	s.Equal(snapshot.Bids[0].Amount, retrieved.Bids[0].Amount)
}

func (s *OrderBookSnapshotsDBTestSuite) TestSaveOrderBookSnapshots() {
	repo := db.NewOrderBookSnapshotsDB(s.conn)
	ctx := s.T().Context()

	snapshots := []*entity.OrderBookSnapshot{
		{
			Pair:      "ETH_USDT",
			Timestamp: 1625097600000,
			Asks:      []entity.OrderBookEntry{{Price: 2000.5, Volume: 1.0, Amount: 2000.5}},
			Bids:      []entity.OrderBookEntry{{Price: 1999.5, Volume: 1.0, Amount: 1999.5}},
		},
		{
			Pair:      "ETH_USDT",
			Timestamp: 1625097601000,
			Asks:      []entity.OrderBookEntry{{Price: 2001.5, Volume: 1.0, Amount: 2001.5}},
			Bids:      []entity.OrderBookEntry{{Price: 2000.5, Volume: 1.0, Amount: 2000.5}},
		},
	}

	err := repo.SaveOrderBookSnapshots(ctx, snapshots)
	s.NoError(err)

	retrieved, err := repo.GetOrderBookSnapshot(ctx, "ETH_USDT")
	s.NoError(err)
	s.NotNil(retrieved)
	s.Equal(int64(1625097601000), retrieved.Timestamp) // should be the latest
}

func (s *OrderBookSnapshotsDBTestSuite) SetupTest() {
	s.container, s.conn, s.dsn = db.SetupPGFlow(&s.Suite, s.migrationsPath)
}

func (s *OrderBookSnapshotsDBTestSuite) TearDownTest() {
	db.TearDownPGFlow(&s.Suite, s.container, s.conn, s.dsn, s.migrationsPath)
}
