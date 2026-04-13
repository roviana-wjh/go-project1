package data

import (
	"context"
	"fmt"

	"review-B/internal/conf"

	v1 "review-B/api/api/review"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewBusinessRepo, NewReviewClient)

// Data .
type Data struct {
	// TODO wrapped database client
	//嵌入一个grpc的client，通过client去调用review-service的服务
	rc  v1.ReviewClient
	log *log.Helper
}

// NewData .
func NewData(c *conf.Data, rc v1.ReviewClient, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.Info("closing the data resources")
	}
	return &Data{
		rc:  rc,
		log: log.NewHelper(logger),
	}, cleanup, nil
}

// NewReviewClient dials review-service via Consul discovery (plaintext; use TLS in production).
func NewReviewClient(c *conf.Data, d registry.Discovery) (v1.ReviewClient, error) {
	if d == nil {
		return nil, fmt.Errorf("registry discovery is required")
	}
	if c == nil || c.GetReview() == nil || c.GetReview().GetService() == "" {
		return nil, fmt.Errorf("data.review.service is required (Consul-registered Kratos service name)")
	}
	ep := "discovery:///" + c.GetReview().GetService()
	conn, err := kgrpc.DialInsecure(
		context.Background(),
		kgrpc.WithEndpoint(ep),
		kgrpc.WithDiscovery(d),
	)
	if err != nil {
		return nil, err
	}
	return v1.NewReviewClient(conn), nil
}
