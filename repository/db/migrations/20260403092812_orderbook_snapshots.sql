-- +goose Up
-- +goose StatementBegin

CREATE TABLE orderbook_snapshots (
    id              BIGSERIAL PRIMARY KEY,
    pair            VARCHAR(20) NOT NULL,
    timestamp_ms    BIGINT NOT NULL,
    payload         JSONB NOT NULL
);

CREATE INDEX idx_ob_snapshots_timestamp ON orderbook_snapshots (pair, timestamp_ms DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS orderbook_snapshots;

-- +goose StatementEnd
