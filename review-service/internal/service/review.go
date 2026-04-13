package service

import (
	"context"
	"fmt"

	pb "review-service/api/review/v1"
	"review-service/internal/biz"
	"review-service/internal/data/model"
)

type ReviewService struct {
	pb.UnimplementedReviewServer
	uc *biz.ReviewUsecase
}

func NewReviewService(uc *biz.ReviewUsecase) *ReviewService {
	return &ReviewService{
		uc: uc,
	}
}

func reviewInfoToListItem(m *model.ReviewInfo) *pb.ReviewListItem {
	if m == nil {
		return nil
	}
	return &pb.ReviewListItem{
		ReviewID:     m.ReviewID,
		UserID:       m.UserID,
		OrderID:      m.OrderID,
		Score:        m.Score,
		ServiceScore: m.ServiceScore,
		ExpressScore: m.ExpressScore,
		Content:      m.Content,
		PicInfo:      m.PicInfo,
		VideoInfo:    m.VideoInfo,
		Status:       m.Status,
		HasReply:     m.HasReply,
		CreateAt:     m.CreateAt.UnixMilli(),
	}
}

func (s *ReviewService) CreateReview(ctx context.Context, req *pb.CreateReviewRequest) (*pb.CreateReviewReply, error) {
	if err := req.Validate(); err != nil {
		return nil, pb.ErrorInvalidParameter("%v", err)
	}
	anonymous := int32(0)
	if req.Anonymous {
		anonymous = 1
	}
	review, err := s.uc.CreateReview(ctx, &model.ReviewInfo{
		UserID:       req.UserID,
		OrderID:      req.OrderID,
		Score:        req.Score,
		ServiceScore: req.ServiceScore,
		ExpressScore: req.ExpressScore,
		Content:      req.Content,
		PicInfo:      req.PicInfo,
		VideoInfo:    req.VideoInfo,
		Status:       biz.ReviewStatusPending,
		Anonymous:    anonymous,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateReviewReply{ReviewID: review.ReviewID}, nil
}

func (s *ReviewService) UpdateReview(ctx context.Context, req *pb.UpdateReviewRequest) (*pb.UpdateReviewReply, error) {
	if err := req.Validate(); err != nil {
		return nil, pb.ErrorInvalidParameter("%v", err)
	}
	fmt.Println("UpdateReview：", req)
	if _, err := s.uc.UpdateReview(ctx, &biz.UpdateReviewParams{
		ReviewID:       req.ReviewID,
		OperatorUserID: req.UserID,
		Score:          req.Score,
		ServiceScore:   req.ServiceScore,
		ExpressScore:   req.ExpressScore,
		Content:        req.Content,
		PicInfo:        req.PicInfo,
		VideoInfo:      req.VideoInfo,
	}); err != nil {
		return nil, err
	}
	return &pb.UpdateReviewReply{}, nil
}

func (s *ReviewService) DeleteReview(ctx context.Context, req *pb.DeleteReviewRequest) (*pb.DeleteReviewReply, error) {
	if err := req.Validate(); err != nil {
		return nil, pb.ErrorInvalidParameter("%v", err)
	}
	fmt.Println("DeleteReview：", req)
	err := s.uc.DeleteReview(ctx, &biz.DeleteReviewParams{ReviewID: req.ReviewID, OperatorUserID: req.UserID})
	if err != nil {
		return nil, err
	}
	return &pb.DeleteReviewReply{}, nil
}

func (s *ReviewService) GetReview(ctx context.Context, req *pb.GetReviewRequest) (*pb.GetReviewReply, error) {
	if err := req.Validate(); err != nil {
		return nil, pb.ErrorInvalidParameter("%v", err)
	}
	fmt.Println("GetReview：", req)
	review, err := s.uc.GetReview(ctx, req.ReviewID)
	if err != nil {
		return nil, err
	}
	return &pb.GetReviewReply{Item: reviewInfoToListItem(review)}, nil
}

func (s *ReviewService) ListReview(ctx context.Context, req *pb.ListReviewRequest) (*pb.ListReviewReply, error) {
	if err := req.Validate(); err != nil {
		return nil, pb.ErrorInvalidParameter("%v", err)
	}
	fmt.Println("ListReview：", req)
	reviews, total, err := s.uc.ListReview(ctx, &biz.ReviewListOrderParams{
		OrderID:  req.OrderID,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*pb.ReviewListItem, 0, len(reviews))
	for _, r := range reviews {
		list = append(list, reviewInfoToListItem(r))
	}
	return &pb.ListReviewReply{List: list, Total: total}, nil
}

func (s *ReviewService) AuditReview(ctx context.Context, req *pb.AuditReviewRequest) (*pb.AuditReviewReply, error) {
	if err := req.Validate(); err != nil {
		return nil, pb.ErrorInvalidParameter("%v", err)
	}
	fmt.Println("AuditReview：", req)
	review, err := s.uc.AuditReview(ctx, &biz.AuditReviewParams{
		ReviewID: req.ReviewID,
		Result:   req.Result,
		Remark:   req.Remark,
		Operator: req.Operator,
	})
	if err != nil {
		return nil, err
	}
	return &pb.AuditReviewReply{ReviewID: review.ReviewID, Status: review.Status}, nil
}

func (s *ReviewService) ReplyReview(ctx context.Context, req *pb.ReplyReviewRequest) (*pb.ReplyReviewReply, error) {
	if err := req.Validate(); err != nil {
		return nil, pb.ErrorInvalidParameter("%v", err)
	}
	fmt.Println("ReplyReview：", req)
	reply, err := s.uc.CreateReply(ctx, &model.ReviewReplyInfo{
		ReviewID:  req.ReviewID,
		StoreID:   req.StoreID,
		Content:   req.Content,
		PicInfo:   req.PicInfo,
		VideoInfo: req.VideoInfo,
		ExtJSON:   req.ExtJSON,
		CtrlJSON:  req.CtrlJSON,
	})
	if err != nil {
		return nil, err
	}
	return &pb.ReplyReviewReply{ReplyID: reply.ReplyID}, nil
}

func (s *ReviewService) AppealReview(ctx context.Context, req *pb.AppealReviewRequest) (*pb.AppealReviewReply, error) {
	if err := req.Validate(); err != nil {
		return nil, pb.ErrorInvalidParameter("%v", err)
	}
	fmt.Println("AppealReview：", req)
	appeal, err := s.uc.AppealReview(ctx, &biz.AppealReviewParams{
		UserID:   req.UserID,
		ReviewID: req.ReviewID,
		Reason:   req.Reason,
		PicInfo:  req.PicInfo,
	})
	if err != nil {
		return nil, err
	}
	return &pb.AppealReviewReply{AppealID: appeal.AppealID}, nil
}

func (s *ReviewService) AuditAppeal(ctx context.Context, req *pb.AuditAppealRequest) (*pb.AuditAppealReply, error) {
	if err := req.Validate(); err != nil {
		return nil, pb.ErrorInvalidParameter("%v", err)
	}
	fmt.Println("AuditAppeal：", req)
	_, err := s.uc.AuditAppeal(ctx, &biz.AuditAppealParams{
		AppealID: req.AppealID,
		Result:   req.Result,
		Remark:   req.Remark,
		Operator: req.Operator,
	})
	if err != nil {
		return nil, err
	}
	return &pb.AuditAppealReply{}, nil
}

func (s *ReviewService) ListReviewByUseId(ctx context.Context, req *pb.ListReviewByUseIdRequest) (*pb.ListReviewReply, error) {
	if err := req.Validate(); err != nil {
		return nil, pb.ErrorInvalidParameter("%v", err)
	}
	fmt.Println("ListReviewByUseId：", req)
	reviews, total, err := s.uc.ListReviewByUseId(ctx, &biz.ReviewListUserParams{
		UserID:   req.UserID,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*pb.ReviewListItem, 0, len(reviews))
	for _, r := range reviews {
		list = append(list, reviewInfoToListItem(r))
	}
	return &pb.ListReviewReply{List: list, Total: total}, nil
}

func (s *ReviewService) ListReviewByStoreId(ctx context.Context, req *pb.ListReviewByStoreIdRequest) (*pb.ListReviewByStoreIdReply, error) {
	if err := req.Validate(); err != nil {
		return nil, pb.ErrorInvalidParameter("%v", err)
	}
	reviews, total, err := s.uc.ListReviewByStoreId(ctx, &biz.ReviewListStoreParams{
		StoreID:  req.StoreID,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*pb.ReviewListItem, 0, len(reviews))
	for _, r := range reviews {
		list = append(list, reviewInfoToListItem(r))
	}
	return &pb.ListReviewByStoreIdReply{List: list, Total: total}, nil
}