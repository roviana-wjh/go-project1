package biz

import (
	"context"

	reviewv1 "review-service/api/review/v1"
)

// ReviewForward 由 data 层实现：封装对 review-service 的 gRPC 调用。
type ReviewForward interface {
	AuditReview(ctx context.Context, in *reviewv1.AuditReviewRequest) (*reviewv1.AuditReviewReply, error)
	AuditAppeal(ctx context.Context, in *reviewv1.AuditAppealRequest) (*reviewv1.AuditAppealReply, error)
	ListPendingReviews(ctx context.Context, in *reviewv1.ListPendingReviewsRequest) (*reviewv1.ListReviewReply, error)
	ListPendingAppeals(ctx context.Context, in *reviewv1.ListPendingAppealsRequest) (*reviewv1.ListPendingAppealsReply, error)
}
