package data

import (
	"context"
	"fmt"

	"review-O/internal/conf"
	reviewv1 "review-service/api/review/v1"

	"github.com/go-kratos/kratos/v2/log"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/google/wire"
	"google.golang.org/grpc"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewReviewForwardRepo)

// Data holds outbound gRPC clients.
type Data struct {
	Review reviewv1.ReviewClient
	conn   *grpc.ClientConn
}

// NewData dials review-service: Consul discovery when registry + data.review.service are set; otherwise data.review.endpoint (plaintext).
func NewData(c *conf.Data, reg *conf.Registry) (*Data, func(), error) {
	rv := c.GetReview()
	if rv == nil {
		return nil, nil, fmt.Errorf("data.review is required")
	}
	ctx := context.Background()
	var conn *grpc.ClientConn
	var err error

	useConsul := reg != nil && reg.GetConsul() != nil && rv.GetService() != ""
	if useConsul {
		disc := newConsulDiscovery(reg)
		ep := "discovery:///" + rv.GetService()
		conn, err = kgrpc.DialInsecure(ctx,
			kgrpc.WithEndpoint(ep),
			kgrpc.WithDiscovery(disc),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("dial review-service via Consul %q: %w", ep, err)
		}
	} else if rv.GetEndpoint() != "" {
		conn, err = kgrpc.DialInsecure(ctx, kgrpc.WithEndpoint(rv.GetEndpoint()))
		if err != nil {
			return nil, nil, fmt.Errorf("dial review-service direct: %w", err)
		}
	} else {
		return nil, nil, fmt.Errorf("set data.review.service + registry.consul for Consul, or data.review.endpoint for direct gRPC")
	}

	d := &Data{
		Review: reviewv1.NewReviewClient(conn),
		conn:   conn,
	}
	cleanup := func() {
		_ = conn.Close()
		log.Info("closing the data resources")
	}
	return d, cleanup, nil
}
