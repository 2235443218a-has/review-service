package biz

import (
	"context"
	v1 "review-service/api/review/v1"
	"review-service/internal/data/model"
	snowflake "review-service/pkg"

	"github.com/go-kratos/kratos/v2/log"
)

type ReviewRepo interface {
	// add methods here
	SaveReview(context.Context, *model.ReviewInfo) (*model.ReviewInfo, error)
	GetReviewByOrderID(context.Context, int64) ([]*model.ReviewInfo, error)
}
type ReviewUsecase struct {
	// add fields here
	repo ReviewRepo
	log  *log.Helper
}

func NewReviewUsecase(repo ReviewRepo, logger log.Logger) *ReviewUsecase {
	return &ReviewUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// CreateReview creates a new review,实现业务逻辑的地方
// service层调用
func (uc *ReviewUsecase) CreateReview(ctx context.Context, review *model.ReviewInfo) (*model.ReviewInfo, error) {
	//数据校验 业务校验 已经评价过不能再评价
	reviews, err := uc.repo.GetReviewByOrderID(ctx, review.OrderID)
	if err != nil {
		return nil, v1.ErrorDbFailed("查询数据库失败")
	}
	if len(reviews) > 0 {
		return nil, v1.ErrorOrderReviewed("订单已评价%d", review.OrderID)
	}
	//生成reviewID（雪花算法）
	review.ReviewID = snowflake.GenID()
	//查询订单和商品快照
	//查询订单和商品快照
	return uc.repo.SaveReview(ctx, review)
}
