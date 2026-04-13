package service

import (
	"context"

	pb "review-B/api/api/business"
	"review-B/internal/biz"
)

type BusinessService struct {
	pb.UnimplementedBusinessServer
	uc *biz.BusinessUsecase
}

func NewBusinessService(uc *biz.BusinessUsecase) *BusinessService {
	return &BusinessService{uc: uc}
}

func (s *BusinessService) ReplyReview(ctx context.Context, req *pb.ReplyReviewRequest) (*pb.ReplyReviewReply, error) {
	replyID, err := s.uc.CreateReply(ctx, &biz.ReplyParams{
		ReviewID: req.GetReviewID(),
		StoreID: req.GetStoreID(),
		Content: req.GetContent(),
		PicInfo: req.GetPicInfo(),
		VideoInfo: req.GetVideoInfo(),
		ExtJSON: req.GetExtJSON(),
		CtrlJSON: req.GetCtrlJSON(),
	})
	if err!=nil{
		return nil,err
	}
	return &pb.ReplyReviewReply{ReplyID: replyID}, nil
}
