package service

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"github.com/pentops/grpc.go/protovalidatemw"
	"github.com/pentops/log.go/grpc_log"
	"github.com/pentops/log.go/log"
	"github.com/pentops/realms/j5auth"
	"google.golang.org/grpc"
)

func GRPCMiddleware() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		grpc_log.UnaryServerInterceptor(log.DefaultContext, log.DefaultTrace, log.DefaultLogger),
		j5auth.GRPCMiddleware,
		protovalidatemw.UnaryServerInterceptor(),
	}
}

type DBConfig struct {
	URL string `env:"POSTGRES_URL"`
}

func (cfg *DBConfig) OpenDatabase(ctx context.Context) (*sql.DB, error) {

	db, err := sql.Open("postgres", cfg.URL)
	if err != nil {
		return nil, err
	}

	// Default is unlimited connections, use a cap to prevent hammering the database if it's the bottleneck.
	// 10 was selected as a conservative number and will likely be revised later.
	db.SetMaxOpenConns(10)

	for {
		if err := db.Ping(); err != nil {
			log.WithError(ctx, err).Error("pinging PG")
			time.Sleep(time.Second)
			continue
		}
		break
	}

	return db, nil
}
