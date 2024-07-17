package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type ReplyParam struct {
	ReviewID  int64
	StoreID   int64
	Content   string
	PicInfo   string
	VideoInfo string
}

type AppealParam struct {
	ReviewID  int64
	StoreID   int64
	Reason    string
	Content   string
	PicInfo   string
	VideoInfo string
}

type BusinessRepo interface {
	Reply(context.Context, *ReplyParam) (int64, error)   //商家对用户的评价进行回复
	Appeal(context.Context, *AppealParam) (int64, error) //商家对用户的违规评价进行申诉
}

type BusinessUsecase struct {
	log  *log.Helper
	repo BusinessRepo
}

func NewBusinessUsecase(logger log.Logger, repo BusinessRepo) *BusinessUsecase {
	return &BusinessUsecase{
		log:  log.NewHelper(logger),
		repo: repo,
	}
}

// CreateReply 创建用户评价的回复,需要调用review-service的RPC服务
func (uc *BusinessUsecase) CreateReply(ctx context.Context, param *ReplyParam) (int64, error) {
	uc.log.WithContext(ctx).Debugf("[biz] CreateReply param:%v", param)
	replyID, err := uc.repo.Reply(ctx, param)
	if err != nil {
		uc.log.Errorf("CreateReply出错,err:%v", err)
		return 0, err
	}
	return replyID, nil
}

// CreateAppeal 对用户的违规评价进行申诉,需要调用review-service的RPC服务
func (uc *BusinessUsecase) CreateAppeal(ctx context.Context, param *AppealParam) (int64, error) {
	uc.log.WithContext(ctx).Debugf("[biz] AppealReview param:%v", param)
	appealID, err := uc.repo.Appeal(ctx, param)
	return appealID, err
}
