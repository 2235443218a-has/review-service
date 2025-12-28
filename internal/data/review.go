package data

import (
	"context"
	"errors"
	v1 "review-service/api/review/v1"
	"review-service/internal/biz"
	"review-service/internal/data/model"
	"review-service/internal/data/query"
	snowflake "review-service/pkg"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
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
	review, err := r.data.query.ReviewInfo.WithContext(ctx).
		Where(r.data.query.ReviewInfo.ReviewID.Eq(rpy.ReviewID)).First()
	if err != nil {
		// 只有明确是 "没找到记录" 时，才报 404
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, v1.ErrorReviewNotFound("该评价不存在或已被删除")
		}
		// 其他错误（如连接超时），直接返回 err，让框架报 500
		r.log.Error("数据库出错了")
		return nil, err
	}
	if review.HasReply == 1 {
		r.log.Debug("该评价已回复，不能重复回复")
		return nil, v1.ErrorReplyReview("该评价已回复，不能重复回复%d", rpy.ReviewID)
	}
	// 水平越权鉴定
	//a商家只能回复a商家的评价
	if review.StoreID != rpy.StoreID {
		r.log.Debug("商家越界了")
		return nil, v1.ErrorReplyReview("不能回复其他商家的评价%d", rpy.ReviewID)
	}

	rpy.ReplyID = snowflake.GenID()
	info := &model.ReviewReplyInfo{
		ReplyID:   rpy.ReplyID,
		ReviewID:  rpy.ReviewID,
		StoreID:   rpy.StoreID,
		Content:   rpy.Content,
		PicInfo:   rpy.PicInfo,
		VideoInfo: rpy.VideoInfo,
	}
	//开启事务，记得更新
	err = r.data.query.Transaction(func(tx *query.Query) error {
		if err := tx.ReviewReplyInfo.WithContext(ctx).Create(info); err != nil {
			r.log.Errorf("failed to create reply: %v", err) // 记录日志是个好习惯
			return err
		}
		//更新评价表
		if _, err := tx.ReviewInfo.WithContext(ctx).
			Where(tx.ReviewInfo.ReviewID.Eq(rpy.ReviewID)).
			Update(tx.ReviewInfo.HasReply, 1); err != nil {
			r.log.Errorf("failed to update review has_reply: %v", err) // 记录日志是个好习惯
			return err
		}
		return nil
	})
	if err != nil {
		r.log.Errorf("[SaveReply] transaction failed: reviewID=%d, err=%v", rpy.ReviewID, err)
		return nil, err
	}
	return rpy, nil
}
