package configs

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/vrischmann/envconfig"
)

func GetRatesServiceConfig() (*RatesServiceConfig, error) {
	cfg := RatesServiceConfig{}

	if err := envconfig.Init(&cfg.DB); err != nil {
		return nil, fmt.Errorf("failed to load db config: %w", err)
	}

	if err := envconfig.Init(&cfg.Service); err != nil {
		return nil, fmt.Errorf("failed to load service config: %w", err)
	}

	serialized, err := os.ReadFile(cfg.Service.GrinexPairsConfigsJson)
	if err != nil {
		return nil, fmt.Errorf("failed to read grinex pairs configs json: %w", err)
	}

	pairs := make(map[string]string)
	if err := json.Unmarshal(serialized, &pairs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal grinex pairs configs json: %w", err)
	}

	cfg.GrinexPairs = pairs

	return &cfg, nil
}
