package biz

import (
	"context"

	v1 "review-B/api/api"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

var (
	// ErrUserNotFound is user not found.
	ErrUserNotFound = errors.NotFound(v1.ErrorReason_USER_NOT_FOUND.String(), "user not found")
)

// Greeter is a Greeter model.
type Greeter struct {
	Hello string
}

type ReplyParams struct {
	ReviewID  int64
	StoreID   int64
	Content   string
	PicInfo   string
	VideoInfo string
	ExtJSON   string
	CtrlJSON  string
}

// BusinessRepo is business data access.
type BusinessRepo interface {
	Reply(context.Context, *ReplyParams) (int64, error)
}

// BusinessUsecase is business usecase.
type BusinessUsecase struct {
	repo BusinessRepo
	log  *log.Helper
}

// NewBusinessUsecase new a Business usecase.
func NewBusinessUsecase(repo BusinessRepo, logger log.Logger) *BusinessUsecase {
	return &BusinessUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *BusinessUsecase) CreateReply(ctx context.Context, params *ReplyParams) (int64, error) {
	uc.log.WithContext(ctx).Debugf("[biz] create reply: %+v", params)
	return uc.repo.Reply(ctx, params)
}
