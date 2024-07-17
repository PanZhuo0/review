package service

import (
	"context"

	pb "business/api/business/v1"
	"business/internal/biz"
)

type BusinessService struct {
	pb.UnimplementedBusinessServer
	uc *biz.BusinessUsecase
}

func NewBusinessService(uc *biz.BusinessUsecase) *BusinessService {
	return &BusinessService{uc: uc}
}

// ReplyReview商家回复用户评价业务,需要传入ReviewID、StoreID、回复的内容、图片、视频信息
func (s *BusinessService) ReplyReview(ctx context.Context, req *pb.ReplyReviewRequest) (*pb.ReplyReviewResponse, error) {
	// 调用biz层的方法
	replyID, err := s.uc.CreateReply(ctx, &biz.ReplyParam{
		ReviewID:  req.GetReviewID(),
		StoreID:   req.GetStoreID(),
		Content:   req.GetContent(),
		PicInfo:   req.GetPicInfo(),
		VideoInfo: req.GetVideoInfo(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.ReplyReviewResponse{ReplyId: replyID}, nil
}

// AppealReview 商家对用户的违规评价进行申诉
func (s *BusinessService) AppealReview(ctx context.Context, req *pb.AppealReviewRequest) (*pb.AppealReviewReply, error) {
	AppealID, err := s.uc.CreateAppeal(ctx, &biz.AppealParam{
		ReviewID:  req.GetReviewID(),
		StoreID:   req.GetStoreID(),
		Content:   req.GetContent(),
		Reason:    req.GetReason(),
		PicInfo:   req.GetPicInfo(),
		VideoInfo: req.GetVideoInfo(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.AppealReviewReply{AppealID: AppealID}, nil
}
