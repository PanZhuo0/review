package data

import (
	v1 "business/api/review/v1"
	"business/internal/biz"
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type businessRepo struct {
	// repo中应该有Data、以及LogHelper
	data *Data
	log  *log.Helper
}

func NewBusinessRepo(data *Data, logger log.Logger) biz.BusinessRepo {
	return &businessRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *businessRepo) Reply(ctx context.Context, param *biz.ReplyParam) (int64, error) {
	r.log.WithContext(ctx).Debugf("[data] Reply,param:%v", param)
	// 调用RPC服务,而不需要连接数据库,
	// 之前都是写操作数据库，现在需要的是通过RPC调用
	reply, err := r.data.rc.ReplyReview(ctx, &v1.ReplyReviewRequest{
		ReviewID:  param.ReviewID,
		StoreID:   param.StoreID,
		Content:   param.Content,
		PicInfo:   param.PicInfo,
		VideoInfo: param.VideoInfo,
	})
	if err != nil {
		return 0, err
	}
	return reply.ReplyID, nil
}

func (r *businessRepo) Appeal(ctx context.Context, param *biz.AppealParam) (int64, error) {
	r.log.WithContext(ctx).Infof("[data] Appeal, param:%v", param)
	// 调用RPC服务,而不需要连接数据库,
	// 之前都是写操作数据库，现在需要的是通过RPC调用
	ret, err := r.data.rc.AppealReview(ctx, &v1.AppealReviewRequest{
		ReviewID:  param.ReviewID,
		StoreID:   param.StoreID,
		Reason:    param.Reason,
		Content:   param.Content,
		PicInfo:   param.PicInfo,
		VideoInfo: param.VideoInfo,
	})
	r.log.WithContext(ctx).Debugf("AppealReview return, ret:%v err:%v", ret, err)
	if err != nil {
		return 0, err
	}
	return ret.GetAppealID(), nil
}
