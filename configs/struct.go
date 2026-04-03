package configs

import "github.com/warpgr/test_task/repository/db"

type (
	RatesServiceConfig struct {
		Service     ServiceConfig
		DB          db.Config
		GrinexPairs map[string]string
	}

	ServiceConfig struct {
		Addr                   string `envconfig:"GRPC_ADDRESS" default:":5001"`
		LogLevel               string `envconfig:"LOG_LEVEL" default:"info"`
		HealtechecEndpoint     string `envconfig:"HEALTHCHECK_ENDPOINT" default:"0.0.0.0:8080"`
		GrinexAPIURL           string `envconfig:"GRINEX_API_URL" default:"https://grinex.io/api/v1/"`
		MigrationsDir          string `envconfig:"MIGRATIONS_DIR" default:"./repository/db/migrations"`
		GrinexPairsConfigsJson string `envconfig:"GRINEX_PAIRS_JSON"`
		ServiceName            string `envconfig:"SERVICE_NAME" default:"rates-service"`
		OTLPEndpoint           string `envconfig:"OTLP_ENDPOINT"`
	}
)
