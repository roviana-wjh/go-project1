package biz

import (
	"context"
	"time"

	pb "review-service/api/review/v1"
	"review-service/internal/data/model"
	"review-service/pkg/snowflake"

	"github.com/go-kratos/kratos/v2/log"
)

// 与表 review_info.status 注释一致：10 待审核；20 通过；30 不通过；40 隐藏
const (
	ReviewStatusPending  int32 = 10
	reviewStatusApproved int32 = 20
	reviewStatusRejected int32 = 30
	reviewStatusHidden   int32 = 40
)

type ReviewRepo interface {
	SaveReview(ctx context.Context, m *model.ReviewInfo) (*model.ReviewInfo, error)
	SaveReply(ctx context.Context, m *model.ReviewReplyInfo) (*model.ReviewReplyInfo, error)
	GetReviewByOrderID(ctx context.Context, orderID int64) ([]*model.ReviewInfo, error)
	GetByReviewID(ctx context.Context, reviewID int64) (*model.ReviewInfo, error)
	ListByOrderID(ctx context.Context, orderID int64, page, pageSize int32) ([]*model.ReviewInfo, int64, error)
	UpdateReview(ctx context.Context, m *model.ReviewInfo) error
	DeleteByReviewID(ctx context.Context, reviewID int64) error
}

type ReviewUsecase struct {
	repo ReviewRepo
	log  *log.Helper
}

func NewReviewUsecase(repo ReviewRepo, logger log.Logger) *ReviewUsecase {
	return &ReviewUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *ReviewUsecase) CreateReview(ctx context.Context, m *model.ReviewInfo) (*model.ReviewInfo, error) {
	uc.log.WithContext(ctx).Debugf("[biz] create review: %+v", m)
	reviews, err := uc.repo.GetReviewByOrderID(ctx, m.OrderID)
	if err != nil {
		return nil, err
	}
	if len(reviews) > 0 {
		return nil, pb.ErrorOrderAlreadyReviewed("订单 %d 已评价", m.OrderID)
	}
	m.ReviewID = snowflake.GenID()
	out, err := uc.repo.SaveReview(ctx, m)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (uc *ReviewUsecase) GetReview(ctx context.Context, reviewID int64) (*model.ReviewInfo, error) {
	row, err := uc.repo.GetByReviewID(ctx, reviewID)
	if err != nil {
		return nil, err
	}
	return row, nil
}
func (uc *ReviewUsecase) CreateReply(ctx context.Context, m *model.ReviewReplyInfo) (*model.ReviewReplyInfo, error) {
	uc.log.WithContext(ctx).Debugf("[biz] create reply: %+v", m)
	reply := &model.ReviewReplyInfo{
		ReviewID:  m.ReviewID,
		StoreID:   m.StoreID,
		Content:   m.Content,
		PicInfo:   m.PicInfo,
		VideoInfo: m.VideoInfo,
		ExtJSON:   m.ExtJSON,
		CtrlJSON:  m.CtrlJSON,
		CreateBy:  m.CreateBy,
		UpdateBy:  m.UpdateBy,
		CreateAt:  time.Now(),
		UpdateAt:  time.Now(),
		DeleteAt:  time.Time{},
		Version:   0,
		ReplyID:   snowflake.GenID(),
	}
	_, err := uc.repo.SaveReply(ctx, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
func (uc *ReviewUsecase) ListReview(ctx context.Context, orderID int64, page, pageSize int32) ([]*model.ReviewInfo, int64, error) {
	return uc.repo.ListByOrderID(ctx, orderID, page, pageSize)
}

func (uc *ReviewUsecase) UpdateReview(ctx context.Context, reviewID, operatorUserID int64, score, serviceScore, expressScore int32, content, picInfo, videoInfo string) (*model.ReviewInfo, error) {
	row, err := uc.repo.GetByReviewID(ctx, reviewID)
	if err != nil {
		return nil, err
	}
	if row.UserID != operatorUserID {
		return nil, pb.ErrorForbidden("无权修改他人评价")
	}
	if !canMutateReview(row.Status) {
		return nil, pb.ErrorReviewStatusInvalid("当前状态不可修改评价 (status=%d)", row.Status)
	}
	row.Score = score
	row.ServiceScore = serviceScore
	row.ExpressScore = expressScore
	row.Content = content
	row.PicInfo = picInfo
	row.VideoInfo = videoInfo
	if err := uc.repo.UpdateReview(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (uc *ReviewUsecase) DeleteReview(ctx context.Context, reviewID, operatorUserID int64) error {
	row, err := uc.repo.GetByReviewID(ctx, reviewID)
	if err != nil {
		return err
	}
	if row.UserID != operatorUserID {
		return pb.ErrorForbidden("无权删除他人评价")
	}
	if !canMutateReview(row.Status) {
		return pb.ErrorReviewStatusInvalid("当前状态不可删除评价 (status=%d)", row.Status)
	}
	return uc.repo.DeleteByReviewID(ctx, reviewID)
}

func canMutateReview(status int32) bool {
	return status == ReviewStatusPending
}
