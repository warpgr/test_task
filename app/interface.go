package app

import "context"

type RatesServiceApp interface {
	Init(ctx context.Context) error
	Run(ctx context.Context) <-chan error
	Shutdown(ctx context.Context) error
}
