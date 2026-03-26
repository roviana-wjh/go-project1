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
	return nil, pb.ErrorInternalError("UpdateReview：请在 proto 中补全请求字段后再接 biz 层")
}

func (s *ReviewService) DeleteReview(ctx context.Context, req *pb.DeleteReviewRequest) (*pb.DeleteReviewReply, error) {
	if err := req.Validate(); err != nil {
		return nil, pb.ErrorInvalidParameter("%v", err)
	}
	return nil, pb.ErrorInternalError("DeleteReview：请在 proto 中补全请求字段后再接 biz 层")
}

func (s *ReviewService) GetReview(ctx context.Context, req *pb.GetReviewRequest) (*pb.GetReviewReply, error) {
	if err := req.Validate(); err != nil {
		return nil, pb.ErrorInvalidParameter("%v", err)
	}
	return nil, pb.ErrorInternalError("GetReview：请在 proto 中补全请求字段后再接 biz 层")
}

func (s *ReviewService) ListReview(ctx context.Context, req *pb.ListReviewRequest) (*pb.ListReviewReply, error) {
	if err := req.Validate(); err != nil {
		return nil, pb.ErrorInvalidParameter("%v", err)
	}
	return nil, pb.ErrorInternalError("ListReview：请在 proto 中补全请求字段后再接 biz 层")
}

func (s *ReviewService) AuditReview(ctx context.Context, req *pb.AuditReviewRequest) (*pb.AuditReviewReply, error) {
	if err := req.Validate(); err != nil {
		return nil, pb.ErrorInvalidParameter("%v", err)
	}
	return nil, pb.ErrorInternalError("AuditReview：待实现")
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
	return nil, pb.ErrorInternalError("AppealReview：待实现")
}

func (s *ReviewService) AuditAppeal(ctx context.Context, req *pb.AuditAppealRequest) (*pb.AuditAppealReply, error) {
	if err := req.Validate(); err != nil {
		return nil, pb.ErrorInvalidParameter("%v", err)
	}
	return nil, pb.ErrorInternalError("AuditAppeal：待实现")
}

func (s *ReviewService) ListReviewByUseId(ctx context.Context, req *pb.ListReviewByUseIdRequest) (*pb.ListReviewReply, error) {
	if err := req.Validate(); err != nil {
		return nil, pb.ErrorInvalidParameter("%v", err)
	}
	return nil, pb.ErrorInternalError("ListReviewByUseId：待实现")
}
