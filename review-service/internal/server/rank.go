package server

import (
	"strconv"

	"review-service/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

func registerRankRoutes(srv *khttp.Server, reviewer *service.ReviewService, logger log.Logger) {
	lg := log.NewHelper(logger)
	r := srv.Route("/")
	r.GET("/v1/review/rank/goods/score", func(ctx khttp.Context) error {
		req := ctx.Request()
		query := req.URL.Query()
		page := int32(1)
		pageSize := int32(10)
		if s := query.Get("page"); s != "" {
			if v, err := strconv.ParseInt(s, 10, 32); err == nil && v > 0 {
				page = int32(v)
			}
		}
		if s := query.Get("pageSize"); s != "" {
			if v, err := strconv.ParseInt(s, 10, 32); err == nil && v > 0 && v <= 100 {
				pageSize = int32(v)
			}
		}
		list, total, err := reviewer.ListGoodsScoreRank(ctx, page, pageSize)
		if err != nil {
			lg.WithContext(ctx).Errorf("list goods score rank failed: %v", err)
			return ctx.Result(500, map[string]string{"message": err.Error()})
		}
		return ctx.Result(200, map[string]any{
			"list":  list,
			"total": total,
		})
	})
}
