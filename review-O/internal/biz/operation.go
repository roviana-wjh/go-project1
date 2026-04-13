package biz

import (
	"context"

	opv1 "review-O/api/operation/v1"
	reviewv1 "review-service/api/review/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// OperationUsecase 编排运营侧用例；下游 RPC 由 data.ReviewForward 执行。
type OperationUsecase struct {
	review ReviewForward
	log    *log.Helper
}

// NewOperationUsecase .
func NewOperationUsecase(rp ReviewForward, logger log.Logger) *OperationUsecase {
	if rp == nil {
		panic("ReviewForward is required")
	}
	return &OperationUsecase{
		review: rp,
		log:    log.NewHelper(logger),
	}
}

func (uc *OperationUsecase) AuditReview(ctx context.Context, req *opv1.AuditReviewRequest) (*opv1.AuditReviewReply, error) {
	uc.log.WithContext(ctx).Debugf("[biz] AuditReview reviewID=%d", req.GetReviewID())
	out, err := uc.review.AuditReview(ctx, &reviewv1.AuditReviewRequest{
		ReviewID: req.GetReviewID(),
		Result:   req.GetResult(),
		Remark:   req.GetRemark(),
		Operator: req.GetOperator(),
	})
	if err != nil {
		return nil, err
	}
	return &opv1.AuditReviewReply{
		ReviewID: out.GetReviewID(),
		Status:   out.GetStatus(),
	}, nil
}

func (uc *OperationUsecase) AuditAppeal(ctx context.Context, req *opv1.AuditAppealRequest) (*opv1.AuditAppealReply, error) {
	uc.log.WithContext(ctx).Debugf("[biz] AuditAppeal appealID=%d", req.GetAppealID())
	_, err := uc.review.AuditAppeal(ctx, &reviewv1.AuditAppealRequest{
		AppealID: req.GetAppealID(),
		Result:   req.GetResult(),
		Remark:   req.GetRemark(),
		Operator: req.GetOperator(),
	})
	if err != nil {
		return nil, err
	}
	return &opv1.AuditAppealReply{}, nil
}
