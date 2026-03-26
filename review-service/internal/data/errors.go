package data

import (
	stderrors "errors"

	pb "review-service/api/review/v1"

	"gorm.io/gorm"
)

func wrapReviewDB(op string, err error) error {
	if err == nil {
		return nil
	}
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return pb.ErrorReviewNotFound("%s: 记录不存在", op)
	}
	return pb.ErrorDatabaseError("%s: %v", op, err)
}
