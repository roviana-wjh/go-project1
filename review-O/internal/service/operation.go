package service

import (
	"context"

	pb "review-O/api/operation/v1"
	"review-O/internal/biz"

	kerrors "github.com/go-kratos/kratos/v2/errors"
)

// OperationService exposes HTTP/gRPC for operators.
type OperationService struct {
	pb.UnimplementedOperationServer
	uc *biz.OperationUsecase
}

// NewOperationService .
func NewOperationService(uc *biz.OperationUsecase) *OperationService {
	return &OperationService{uc: uc}
}

func (s *OperationService) AuditReview(ctx context.Context, req *pb.AuditReviewRequest) (*pb.AuditReviewReply, error) {
	if err := req.Validate(); err != nil {
		return nil, kerrors.BadRequest("VALIDATION_FAILED", err.Error())
	}
	return s.uc.AuditReview(ctx, req)
}

func (s *OperationService) AuditAppeal(ctx context.Context, req *pb.AuditAppealRequest) (*pb.AuditAppealReply, error) {
	if err := req.Validate(); err != nil {
		return nil, kerrors.BadRequest("VALIDATION_FAILED", err.Error())
	}
	return s.uc.AuditAppeal(ctx, req)
}
