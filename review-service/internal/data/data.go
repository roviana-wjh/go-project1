package data

import (
	"fmt"
	"review-service/internal/conf"
	"review-service/internal/data/query"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewReviewRepo, NewDB, NewES, NewRedis)

// Data .
type Data struct {
	// TODO wrapped database client
	q   *query.Query
	log *log.Helper
	es  *elasticsearch.TypedClient
	rdb *redis.Client
}

// NewData .
func NewData(db *gorm.DB, es *elasticsearch.TypedClient, rdb *redis.Client, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.Info("closing the data resources")
	}
	//为gen生成的代码设置db
	query.SetDefault(db)
	return &Data{q: query.Q, es: es, rdb: rdb, log: log.NewHelper(logger)}, cleanup, nil
}

func NewDB(c *conf.Data) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})
}

func NewES(c *conf.Elasticsearch) (*elasticsearch.TypedClient, error) {
	if c == nil {
		return nil, fmt.Errorf("elasticsearch config is nil")
	}
	addrs := c.GetAddresses()
	if len(addrs) == 0 {
		return nil, fmt.Errorf("elasticsearch addresses is empty")
	}
	return elasticsearch.NewTypedClient(elasticsearch.Config{Addresses: addrs})
}

func NewRedis(c *conf.Data) (*redis.Client, error) {
	return redis.NewClient(&redis.Options{
		Addr:     c.Redis.Addr,
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
		ReadTimeout: c.Redis.ReadTimeout.AsDuration(),
	}), nil
}
