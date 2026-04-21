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

func toPendingReviewItem(in *reviewv1.ReviewListItem) *opv1.PendingReviewItem {
	if in == nil {
		return nil
	}
	return &opv1.PendingReviewItem{
		ReviewID:     in.GetReviewID(),
		UserID:       in.GetUserID(),
		OrderID:      in.GetOrderID(),
		Score:        in.GetScore(),
		ServiceScore: in.GetServiceScore(),
		ExpressScore: in.GetExpressScore(),
		Content:      in.GetContent(),
		PicInfo:      in.GetPicInfo(),
		VideoInfo:    in.GetVideoInfo(),
		Status:       in.GetStatus(),
		HasReply:     in.GetHasReply(),
		CreateAt:     in.GetCreateAt(),
	}
}

func toPendingAppealItem(in *reviewv1.AppealListItem) *opv1.PendingAppealItem {
	if in == nil {
		return nil
	}
	return &opv1.PendingAppealItem{
		AppealID:  in.GetAppealID(),
		ReviewID:  in.GetReviewID(),
		StoreID:   in.GetStoreID(),
		Status:    in.GetStatus(),
		Reason:    in.GetReason(),
		Content:   in.GetContent(),
		PicInfo:   in.GetPicInfo(),
		VideoInfo: in.GetVideoInfo(),
		OpRemarks: in.GetOpRemarks(),
		OpUser:    in.GetOpUser(),
		CreateAt:  in.GetCreateAt(),
	}
}

func (uc *OperationUsecase) ListPendingReviews(ctx context.Context, req *opv1.ListPendingReviewsRequest) (*opv1.ListPendingReviewsReply, error) {
	uc.log.WithContext(ctx).Debugf("[biz] ListPendingReviews page=%d pageSize=%d", req.GetPage(), req.GetPageSize())
	out, err := uc.review.ListPendingReviews(ctx, &reviewv1.ListPendingReviewsRequest{
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	})
	if err != nil {
		return nil, err
	}
	list := make([]*opv1.PendingReviewItem, 0, len(out.GetList()))
	for _, item := range out.GetList() {
		list = append(list, toPendingReviewItem(item))
	}
	return &opv1.ListPendingReviewsReply{
		List:  list,
		Total: out.GetTotal(),
	}, nil
}

func (uc *OperationUsecase) ListPendingAppeals(ctx context.Context, req *opv1.ListPendingAppealsRequest) (*opv1.ListPendingAppealsReply, error) {
	uc.log.WithContext(ctx).Debugf("[biz] ListPendingAppeals page=%d pageSize=%d", req.GetPage(), req.GetPageSize())
	out, err := uc.review.ListPendingAppeals(ctx, &reviewv1.ListPendingAppealsRequest{
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	})
	if err != nil {
		return nil, err
	}
	list := make([]*opv1.PendingAppealItem, 0, len(out.GetList()))
	for _, item := range out.GetList() {
		list = append(list, toPendingAppealItem(item))
	}
	return &opv1.ListPendingAppealsReply{
		List:  list,
		Total: out.GetTotal(),
	}, nil
}
