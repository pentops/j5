package service

import (
	_ "github.com/lib/pq"
	"github.com/pentops/grpc.go/protovalidatemw"
	"github.com/pentops/grpc.go/versionmw"
	"github.com/pentops/log.go/grpc_log"
	"github.com/pentops/log.go/log"
	"github.com/pentops/realms/j5auth"
	"google.golang.org/grpc"
)

func GRPCMiddleware(version string) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		grpc_log.UnaryServerInterceptor(log.DefaultContext, log.DefaultTrace, log.DefaultLogger),
		j5auth.GRPCMiddleware,
		protovalidatemw.UnaryServerInterceptor(),
		versionmw.UnaryServerInterceptor(version),
	}
}
