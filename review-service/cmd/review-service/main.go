package main

import (
	"flag"
	"os"

	"review-service/internal/conf"
	"review-service/pkg/snowflake"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

var (
	Name     string = "review-service"
	Version  string = "1.0.0"
	flagconf string
	id, _    = os.Hostname()
)

func defaultConfDir() string {
	// Order matters: from repo root (e.g. kratos run picking cmd), ../../configs may
	// resolve to an unrelated directory; prefer this service's configs first.
	for _, p := range []string{
		"configs",
		"../configs",
		"review-service/configs",
		"../review-service/configs",
		"../../configs",
	} {
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			return p
		}
	}
	return "review-service/configs"
}

func init() {
	flag.StringVar(&flagconf, "conf", defaultConfDir(), "config directory (yaml inside)")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, r registry.Registrar) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(gs, hs),
		kratos.Registrar(r),
	)
}

func main() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	c := config.New(config.WithSource(file.NewSource(flagconf)))
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	if sf := bc.GetSnowflake(); sf != nil {
		if err := snowflake.Init(sf.GetStartTime(), sf.GetMachineId()); err != nil {
			panic(err)
		}
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Registry, bc.GetElasticsearch(), logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
