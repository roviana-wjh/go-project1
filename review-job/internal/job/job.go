package job

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// NewLoggerHelper wraps log.NewHelper so Wire can inject log.Logger without
// binding variadic ...log.Option providers.
func NewLoggerHelper(logger log.Logger) *log.Helper {
	return log.NewHelper(logger)
}

var ProviderSet = wire.NewSet(
	NewKafkaReader,
	NewESClient,
	NewLoggerHelper,
	NewJobWorker,
)
