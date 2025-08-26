package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pentops/flowtest"
	"github.com/pentops/j5/gen/j5/registry/v1/registry_tpb"
	"github.com/pentops/j5/internal/gen/j5/registry/v1/registry_spb"
	"github.com/pentops/j5/internal/registry/anyfs"
	"github.com/pentops/j5/internal/registry/gomodproxy"
	"github.com/pentops/j5/internal/registry/integration/mocks"
	"github.com/pentops/j5/internal/registry/packagestore"
	"github.com/pentops/j5/internal/registry/service"
	"github.com/pentops/log.go/log"
	"github.com/pentops/o5-messaging/outbox/outboxtest"
	"github.com/pentops/pgtest.go/pgtest"
	"github.com/pentops/sqrlx.go/sqrlx"
)

type Universe struct {
	Outbox *outboxtest.OutboxAsserter

	BuilderRequest   registry_tpb.BuilderRequestTopicClient
	RegistryDownload registry_spb.DownloadServiceClient

	PackageStore *packagestore.PackageStore

	Github *mocks.GithubMock

	HTTPHandler http.Handler
}

func NewUniverse(t *testing.T) (*flowtest.Stepper[*testing.T], *Universe) {
	name := t.Name()
	stepper := flowtest.NewStepper[*testing.T](name)
	uu := &Universe{}

	stepper.Setup(func(ctx context.Context, t flowtest.Asserter) error {
		log.DefaultLogger = log.NewCallbackLogger(stepper.LevelLog)
		setupUniverse(ctx, t, uu)
		return nil
	})

	stepper.PostStepHook(func(ctx context.Context, t flowtest.Asserter) error {
		uu.Outbox.AssertEmpty(t)
		return nil
	})

	return stepper, uu
}

const TestVersion = "test-version"

// setupUniverse should only be called from the Setup callback, it is effectively
// a method but shouldn't show up there.
func setupUniverse(ctx context.Context, t flowtest.Asserter, uu *Universe) {
	t.Helper()

	conn := pgtest.GetTestDB(t, pgtest.WithDir("../../../ext/db"))
	db := sqrlx.NewPostgres(conn)

	uu.Outbox = outboxtest.NewOutboxAsserter(t, conn)
	uu.Github = mocks.NewGithubMock()

	grpcPair := flowtest.NewGRPCPair(t, service.GRPCMiddleware("test")...)

	tmpfs, err := anyfs.NewTempFS(ctx)
	if err != nil {
		t.Fatalf("failed to create temp fs: %v", err)
	}

	pkgStore, err := packagestore.NewPackageStore(db, tmpfs)
	if err != nil {
		t.Fatalf("failed to create package store: %v", err)
	}

	uu.PackageStore = pkgStore

	uu.HTTPHandler = gomodproxy.Handler(pkgStore)

	registryService := service.NewRegistryService(pkgStore)
	registryService.RegisterGRPC(grpcPair.Server)

	uu.BuilderRequest = registry_tpb.NewBuilderRequestTopicClient(grpcPair.Client)
	uu.RegistryDownload = registry_spb.NewDownloadServiceClient(grpcPair.Client)

	grpcPair.ServeUntilDone(t, ctx)
}

type HTTPResponse struct {
	Body       []byte
	StatusCode int
}

func (uu *Universe) HTTPGet(ctx context.Context, path string) HTTPResponse {
	req := httptest.NewRequest("GET", path, nil)
	req = req.WithContext(ctx)

	res := httptest.NewRecorder()
	uu.HTTPHandler.ServeHTTP(res, req)

	out := HTTPResponse{
		Body:       res.Body.Bytes(),
		StatusCode: res.Code,
	}

	log.WithFields(ctx, map[string]any{
		"status": res.Code,
		//"body":   string(out.Body),
		"path": path,
	}).Info("HTTP GET")

	return out
}
