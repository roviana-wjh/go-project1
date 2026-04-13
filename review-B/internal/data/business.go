package data

import (
	"context"

	reviewv1 "review-B/api/api/review"
	"review-B/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type businessRepo struct {
	data *Data
	log  *log.Helper
}

// NewGreeterRepo .
func NewBusinessRepo(data *Data, logger log.Logger) biz.BusinessRepo {
	return &businessRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *businessRepo) Reply(ctx context.Context, param *biz.ReplyParams) (int64, error) {
	r.log.WithContext(ctx).Debugf("[data] reply: %+v", param)
	reply, err := r.data.rc.ReplyReview(ctx, &reviewv1.ReplyReviewRequest{
		ReviewID:  param.ReviewID,
		StoreID:   param.StoreID,
		Content:   param.Content,
		PicInfo:   param.PicInfo,
		VideoInfo: param.VideoInfo,
		ExtJSON:   param.ExtJSON,
		CtrlJSON:  param.CtrlJSON,
	})

	if err != nil {
		return 0, err
	}
	return reply.GetReplyID(), nil
}

// func (r *greeterRepo) Update(ctx context.Context, g *biz.Greeter) (*biz.Greeter, error) {
// 	return g, nil
// }

// func (r *greeterRepo) FindByID(context.Context, int64) (*biz.Greeter, error) {
// 	return nil, nil
// }

// func (r *greeterRepo) ListByHello(context.Context, string) ([]*biz.Greeter, error) {
// 	return nil, nil
// }

// func (r *greeterRepo) ListAll(context.Context) ([]*biz.Greeter, error) {
// 	return nil, nil
// }
