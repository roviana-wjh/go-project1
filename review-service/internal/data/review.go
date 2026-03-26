package data

import (
	"context"
	"errors"

	"review-service/internal/biz"
	"review-service/internal/data/model"
	"review-service/internal/data/query"

	"github.com/go-kratos/kratos/v2/log"
)

type ReviewRepo struct {
	data *Data
	log  *log.Helper
}

func NewReviewRepo(data *Data, logger log.Logger) biz.ReviewRepo {
	return &ReviewRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *ReviewRepo) SaveReview(ctx context.Context, review *model.ReviewInfo) (*model.ReviewInfo, error) {
	if err := r.data.q.ReviewInfo.WithContext(ctx).Create(review); err != nil {
		return nil, wrapReviewDB("创建评价", err)
	}
	return review, nil
}

func (r *ReviewRepo) GetReviewByOrderID(ctx context.Context, orderID int64) ([]*model.ReviewInfo, error) {
	reviews, err := r.data.q.ReviewInfo.WithContext(ctx).Where(r.data.q.ReviewInfo.OrderID.Eq(orderID)).Find()
	if err != nil {
		return nil, wrapReviewDB("按订单查询评价", err)
	}
	return reviews, nil
}

func (r *ReviewRepo) GetByReviewID(ctx context.Context, reviewID int64) (*model.ReviewInfo, error) {
	row, err := r.data.q.ReviewInfo.WithContext(ctx).Where(r.data.q.ReviewInfo.ReviewID.Eq(reviewID)).First()
	if err != nil {
		return nil, wrapReviewDB("查询评价", err)
	}
	return row, nil
}

func (r *ReviewRepo) ListByOrderID(ctx context.Context, orderID int64, page, pageSize int32) ([]*model.ReviewInfo, int64, error) {
	q := r.data.q.ReviewInfo.WithContext(ctx).Where(r.data.q.ReviewInfo.OrderID.Eq(orderID))
	offset := int((page - 1) * pageSize)
	list, total, err := q.Order(r.data.q.ReviewInfo.CreateAt.Desc()).FindByPage(offset, int(pageSize))
	if err != nil {
		return nil, 0, wrapReviewDB("分页查询评价", err)
	}
	return list, total, nil
}

func (r *ReviewRepo) UpdateReview(ctx context.Context, row *model.ReviewInfo) error {
	if err := r.data.q.ReviewInfo.WithContext(ctx).Save(row); err != nil {
		return wrapReviewDB("更新评价", err)
	}
	return nil
}

func (r *ReviewRepo) DeleteByReviewID(ctx context.Context, reviewID int64) error {
	info, err := r.data.q.ReviewInfo.WithContext(ctx).Where(r.data.q.ReviewInfo.ReviewID.Eq(reviewID)).First()
	if err != nil {
		return wrapReviewDB("删除前查询评价", err)
	}
	if _, err := r.data.q.ReviewInfo.WithContext(ctx).Delete(info); err != nil {
		return wrapReviewDB("删除评价", err)
	}
	return nil
}
func (r *ReviewRepo) SaveReply(ctx context.Context, reply *model.ReviewReplyInfo) (*model.ReviewReplyInfo, error) {
	//数据校验
	//1. 已回复的评价不允许商家再次回复
	review, err := r.data.q.ReviewInfo.WithContext(ctx).Where(r.data.q.ReviewInfo.ReviewID.Eq(reply.ReviewID)).First()
	if err != nil {
		return nil, errors.New("评价不存在")
	}
	if review.HasReply == 1 {
		return nil, errors.New("已回复的评价不允许商家再次回复")
	}

	//2. 水平越权校验 A-B 权限
	if review.StoreID != reply.StoreID {
		return nil, errors.New("无权回复他人评价")
	}
	// 3. 同一事务：写入回复 + 将评价 has_reply 置为 1
	review.HasReply = 1
	if err := r.data.q.Transaction(func(tx *query.Query) error {
		if err := tx.ReviewReplyInfo.WithContext(ctx).Save(reply); err != nil {
			return err
		}
		if err := tx.ReviewInfo.WithContext(ctx).Where(tx.ReviewInfo.ReviewID.Eq(review.ReviewID)).Save(review); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, wrapReviewDB("保存商家回复", err)
	}
	return reply, nil
}
