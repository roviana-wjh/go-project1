package data

import (
	"context"

	"review-O/internal/biz"
	reviewv1 "review-service/api/review/v1"
)

type reviewForwardRepo struct {
	cli reviewv1.ReviewClient
}

var _ biz.ReviewForward = (*reviewForwardRepo)(nil)

// NewReviewForwardRepo 封装 review-service gRPC，供 biz 注入。
func NewReviewForwardRepo(d *Data) biz.ReviewForward {
	if d == nil || d.Review == nil {
		panic("review-service gRPC client is required")
	}
	return &reviewForwardRepo{cli: d.Review}
}

func (r *reviewForwardRepo) AuditReview(ctx context.Context, in *reviewv1.AuditReviewRequest) (*reviewv1.AuditReviewReply, error) {
	return r.cli.AuditReview(ctx, in)
}

func (r *reviewForwardRepo) AuditAppeal(ctx context.Context, in *reviewv1.AuditAppealRequest) (*reviewv1.AuditAppealReply, error) {
	return r.cli.AuditAppeal(ctx, in)
}

func (r *reviewForwardRepo) ListPendingReviews(ctx context.Context, in *reviewv1.ListPendingReviewsRequest) (*reviewv1.ListReviewReply, error) {
	return r.cli.ListPendingReviews(ctx, in)
}

func (r *reviewForwardRepo) ListPendingAppeals(ctx context.Context, in *reviewv1.ListPendingAppealsRequest) (*reviewv1.ListPendingAppealsReply, error) {
	return r.cli.ListPendingAppeals(ctx, in)
}
