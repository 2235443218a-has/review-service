package data

import (
	"context"
	"review-service/internal/biz"
	"review-service/internal/data/model"
	snowflake "review-service/pkg"

	"github.com/go-kratos/kratos/v2/log"
)

type reviewRepo struct {
	data *Data
	log  *log.Helper
}

// NewReviewRepo .
func NewReviewRepo(data *Data, logger log.Logger) biz.ReviewRepo {
	return &reviewRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *reviewRepo) SaveReview(ctx context.Context, review *model.ReviewInfo) (*model.ReviewInfo, error) {
	err := r.data.query.ReviewInfo.WithContext(ctx).Save(review)
	return review, err
}

func (r *reviewRepo) GetReviewByOrderID(ctx context.Context, orderID int64) ([]*model.ReviewInfo, error) {

	reviews, err := r.data.query.ReviewInfo.WithContext(ctx).
		Where(r.data.query.ReviewInfo.OrderID.Eq(orderID)).Find()
	return reviews, err
}
func (r *reviewRepo) SaveReply(ctx context.Context, rpy *biz.ReplyParam) (*biz.ReplyParam, error) {

	//数据校验 已经回复的不能回复，水平越权鉴定
	rpy.ReplyID = snowflake.GenID()
	info := &model.ReviewReplyInfo{
		ReplyID:   rpy.ReplyID,
		ReviewID:  rpy.ReviewID,
		StoreID:   rpy.StoreID,
		Content:   rpy.Content,
		PicInfo:   rpy.PicInfo,
		VideoInfo: rpy.VideoInfo,
	}
	if err := r.data.db.WithContext(ctx).Create(info).Error; err != nil {
		r.log.Errorf("failed to create reply: %v", err) // 记录日志是个好习惯
		return nil, err
	}
	return rpy, nil
}
