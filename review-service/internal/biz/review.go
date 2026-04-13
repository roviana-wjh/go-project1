package biz

import (
	"context"
	"fmt"
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
	ListByOrderID(ctx context.Context, p *ReviewListOrderParams) ([]*model.ReviewInfo, int64, error)
	ListByUseId(ctx context.Context, p *ReviewListUserParams) ([]*model.ReviewInfo, int64, error)
	ListByStoreId(ctx context.Context, p *ReviewListStoreParams) ([]*model.ReviewInfo, int64, error)
	UpdateReview(ctx context.Context, m *model.ReviewInfo) error
	DeleteByReviewID(ctx context.Context, reviewID int64) error
	GetByAppealID(ctx context.Context, appealID int64) (*model.ReviewAppealInfo, error)
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
	reviews, err := uc.repo.GetReviewByOrderID(ctx, m.OrderID)
	if err != nil {
		return nil, err
	}
	if len(reviews) > 0 {
		return nil, pb.ErrorOrderAlreadyReviewed("订单 %d 已评价", m.OrderID)
	}
	uc.log.WithContext(ctx).Debugf("[biz] create review: %+v", m)
	m.ReviewID = snowflake.GenID()
	out, err := uc.repo.SaveReview(ctx, m)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (uc *ReviewUsecase) GetReview(ctx context.Context, reviewID int64) (*model.ReviewInfo, error) {
	uc.log.WithContext(ctx).Debugf("[biz] get review: %d", reviewID)
	row, err := uc.repo.GetByReviewID(ctx, reviewID)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("[biz] get review error: %v", err)
		return nil, err
	}
	uc.log.WithContext(ctx).Debugf("[biz] get review success: %+v", row)
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
		DeleteAt:  nil,
		Version:   0,
		ReplyID:   snowflake.GenID(),
	}
	_, err := uc.repo.SaveReply(ctx, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
func (uc *ReviewUsecase) ListReview(ctx context.Context, p *ReviewListOrderParams) ([]*model.ReviewInfo, int64, error) {
	uc.log.WithContext(ctx).Debugf("[biz] list review: %+v", p)
	return uc.repo.ListByOrderID(ctx, p)
}

func (uc *ReviewUsecase) UpdateReview(ctx context.Context, p *UpdateReviewParams) (*model.ReviewInfo, error) {
	uc.log.WithContext(ctx).Debugf("[biz] update review: %+v", p)
	row, err := uc.repo.GetByReviewID(ctx, p.ReviewID)
	if err != nil {
		return nil, err
	}
	if row.UserID != p.OperatorUserID {
		return nil, pb.ErrorForbidden("无权修改他人评价")
	}
	if !canMutateReview(row.Status) {
		return nil, pb.ErrorReviewStatusInvalid("当前状态不可修改评价 (status=%d)", row.Status)
	}
	row.Score = p.Score
	row.ServiceScore = p.ServiceScore
	row.ExpressScore = p.ExpressScore
	row.Content = p.Content
	row.PicInfo = p.PicInfo
	row.VideoInfo = p.VideoInfo
	if err := uc.repo.UpdateReview(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (uc *ReviewUsecase) DeleteReview(ctx context.Context, p *DeleteReviewParams) error {
	uc.log.WithContext(ctx).Debugf("[biz] delete review: %+v", p)
	row, err := uc.repo.GetByReviewID(ctx, p.ReviewID)
	if err != nil {
		return err
	}
	if row.UserID != p.OperatorUserID {
		return pb.ErrorForbidden("无权删除他人评价")
	}
	if !canMutateReview(row.Status) {
		return pb.ErrorReviewStatusInvalid("当前状态不可删除评价 (status=%d)", row.Status)
	}
	return uc.repo.DeleteByReviewID(ctx, p.ReviewID)
}

func canMutateReview(status int32) bool {
	return status == ReviewStatusPending
}

// AuditReview 运营审核：仅待审核(10)可审；result 1=通过→20，2=驳回→30。operator 为运营账号标识（鉴权应在网关/中间件完成）。
func (uc *ReviewUsecase) AuditReview(ctx context.Context, p *AuditReviewParams) (*model.ReviewInfo, error) {
	uc.log.WithContext(ctx).Debugf("[biz] audit review: %+v", p)
	row, err := uc.repo.GetByReviewID(ctx, p.ReviewID)
	if err != nil {
		return nil, err
	}
	if row.Status != ReviewStatusPending {
		return nil, pb.ErrorReviewStatusInvalid("仅待审核状态可运营审核 (status=%d)", row.Status)
	}
	switch p.Result {
	case 1:
		row.Status = reviewStatusApproved
		if p.Remark != "" {
			row.OpRemarks = p.Remark
		}
	case 2:
		row.Status = reviewStatusRejected
		row.OpReason = p.Remark
	default:
		return nil, pb.ErrorInvalidParameter("result 须为 1(通过) 或 2(驳回)")
	}
	row.OpUser = p.Operator
	row.UpdateAt = time.Now()
	if err := uc.repo.UpdateReview(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (uc *ReviewUsecase) AppealReview(ctx context.Context, p *AppealReviewParams) (*model.ReviewAppealInfo, error) {
	uc.log.WithContext(ctx).Debugf("[biz] appeal review: %+v", p)
	row, err := uc.repo.GetByReviewID(ctx, p.ReviewID)
	if err != nil {
		return nil, err
	}
	if row.UserID != p.UserID {
		return nil, pb.ErrorForbidden("无权申诉他人评价")
	}
	return &model.ReviewAppealInfo{
		AppealID: snowflake.GenID(),
		ReviewID: p.ReviewID,
		StoreID:  row.StoreID,
		Reason:   p.Reason,
		PicInfo:  p.PicInfo,
		CreateBy: fmt.Sprintf("%d", p.UserID),
		UpdateBy: fmt.Sprintf("%d", p.UserID),
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
	}, nil
}

func (uc *ReviewUsecase) AuditAppeal(ctx context.Context, p *AuditAppealParams) (*model.ReviewAppealInfo, error) {
	uc.log.WithContext(ctx).Debugf("[biz] audit appeal: %+v", p)
	row, err := uc.repo.GetByAppealID(ctx, p.AppealID)
	if err != nil {
		return nil, err
	}
	if row.Status != ReviewStatusPending {
		return nil, pb.ErrorReviewStatusInvalid("仅待审核状态可运营审核 (status=%d)", row.Status)
	}
	return &model.ReviewAppealInfo{
		AppealID: p.AppealID,
		ReviewID: row.ReviewID,
		StoreID:  row.StoreID,
		Reason:   row.Reason,
		PicInfo:  row.PicInfo,
		CreateBy: row.CreateBy,
		UpdateBy: p.Operator,
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
		Status:   p.Result,
	}, nil
}

func (uc *ReviewUsecase) ListReviewByUseId(ctx context.Context, p *ReviewListUserParams) ([]*model.ReviewInfo, int64, error) {
	uc.log.WithContext(ctx).Debugf("[biz] list review by user: %+v", p)
	return uc.repo.ListByUseId(ctx, p)
}

func (uc *ReviewUsecase) ListReviewByStoreId(ctx context.Context, p *ReviewListStoreParams) ([]*model.ReviewInfo, int64, error) {
	uc.log.WithContext(ctx).Debugf("[biz] list review by store: %+v", p)
	return uc.repo.ListByStoreId(ctx, p)
}
