package main

import (
	"context"

	"github.com/pentops/log.go/log"
	"gopkg.daemonl.com/envconf"
)

var Version = "dev"

type Config struct {
	Port int `env:"PORT" default:"8080"`
}

func main() {
	ctx := context.Background()
	ctx = log.WithFields(ctx, map[string]interface{}{
		"app":     "jsonapi",
		"version": Version,
	})

	cfg := Config{}
	if err := envconf.Parse(&cfg); err != nil {
		log.Fatal(ctx, err.Error())
	}

	if err := run(ctx, cfg); err != nil {
		log.Fatal(ctx, err.Error())
	}

}

func run(ctx context.Context, cfg Config) error {

	return nil
}
