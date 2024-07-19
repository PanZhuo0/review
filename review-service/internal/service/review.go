package service

import (
	"context"
	"fmt"

	pb "review-service/api/review/v1"
	"review-service/internal/biz"
	"review-service/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
)

type ReviewService struct {
	pb.UnimplementedReviewServer

	// service层需要调用biz层的usecase来实现业务逻辑
	uc *biz.ReviewUsecase
}

func NewReviewService(uc *biz.ReviewUsecase) *ReviewService {
	return &ReviewService{
		uc: uc,
	}
}

/*
CreateReview 创建评价-service层
1.从请求参数req的数据转换到MODEL中，类似JAVA的DTO->VO
2.调用biz层的CreateReview方法
3.返回结果
*/
func (s *ReviewService) CreateReview(ctx context.Context, req *pb.CreateReviewRequest) (*pb.CreateReviewReply, error) {
	// fmt.Printf("[service] createReview,req:%#v", req)
	// 业务参数校验(交给KRATOS,validator完成)
	var annoymous int32
	if req.Anonymous {
		annoymous = 1
	}
	retReview, err := s.uc.CreateReview(ctx, &model.ReviewInfo{ //调用biz层的方法
		OrderID:      req.OrderID,      //订单ID
		UserID:       req.UserID,       //用户ID
		StoreID:      req.StoreID,      //商家ID
		Score:        req.Score,        //评价评分
		ExpressScore: req.ExpressScore, //物流评分
		Content:      req.Content,      //评价内容
		PicInfo:      req.PicInfo,      //照片信息
		VideoInfo:    req.VideoInfo,    //视频信息
		Anonymous:    annoymous,        //是否匿名?
		Status:       0,                //评价的状态
	})
	// Addition:从商户端获取其他信息
	if err != nil {
		log.Error("create review failed,err:", err)
		return nil, err
	}
	// 拼装返回结果
	return &pb.CreateReviewReply{ReviewID: retReview.ReviewID}, nil //返回给调用者记录到数据库中Review的ReviewID
}

// GetReview 获取评价信息,传入参数为ReviewID
func (s *ReviewService) GetReview(ctx context.Context, req *pb.GetReviewRequest) (*pb.GetReviewReply, error) {
	// DTO->PO
	fmt.Printf("[service]GetReview req:%v", req)
	review, err := s.uc.GetReview(ctx, req.ReviewID)
	return &pb.GetReviewReply{Data: &pb.ReviewInfo{
		ReviewID:     review.ReviewID,
		UserID:       review.UserID,
		OrderID:      review.OrderID,
		Score:        review.Score,
		ServiceScore: review.ServiceScore,
		ExpressScore: review.ExpressScore,
		Content:      review.Content,
		PicInfo:      review.PicInfo,
		VideoInfo:    review.VideoInfo,
		Status:       review.Status,
	}}, err
}

// AuditReview 审核用户评论,传入参数为ReviewID、以及运营人员Operator的信息
func (s *ReviewService) AuditReview(ctx context.Context, req *pb.AuditReviewRequest) (*pb.AuditReviewReply, error) {
	fmt.Printf("AuditReview req:%v", req)
	// DTO->PO
	err := s.uc.AuditReview(ctx, &biz.AuditParam{
		ReviewID:  req.ReviewID,
		OpUser:    req.OpUser,
		OpReason:  req.OpReason,
		OpRemarks: *req.OpRemarks,
		Status:    req.Status,
	})
	if err != nil {
		return nil, err
	}
	return &pb.AuditReviewReply{ReviewID: req.ReviewID, Status: req.Status}, nil
}

// ReplyReview 商家回应用户的评论,传入参数为ReviewID、StoreID、回复的内容、图片、视频等信息
func (s *ReviewService) ReplyReview(ctx context.Context, req *pb.ReplyReviewRequest) (*pb.ReplyReviewReply, error) {
	fmt.Printf("[service] ReplyReview req:%v", req)
	reply, err := s.uc.CreateReply(ctx, &biz.ReplyParam{
		ReviewID:  req.ReviewID,
		StoreID:   req.StoreID,
		Content:   req.Content,
		PicInfo:   req.PicInfo,
		VideoInfo: req.VideoInfo,
	})
	if err != nil {
		return nil, err
	}
	return &pb.ReplyReviewReply{ReplyID: reply.ReplyID}, nil
}

// AppealReview 商家对用户的评论进行申诉,传入参数为ReviewID、StoreID、申述的原因、内容、图片、视频等依据
func (s *ReviewService) AppealReview(ctx context.Context, req *pb.AppealReviewRequest) (*pb.AppealReviewReply, error) {
	fmt.Printf("[service] AppealReview req:%v\n", req)
	appeal, err := s.uc.AppealReview(ctx, &biz.AppealParam{
		ReviewID:  req.GetReviewID(),
		StoreID:   req.GetStoreID(),
		Reason:    req.GetReason(),
		Content:   req.GetContent(),
		PicInfo:   req.GetPicInfo(),
		VideoInfo: req.GetVideoInfo(),
	})
	if err != nil {
		return nil, err
	}
	fmt.Printf("[service] AppealReview ret:%v err:%v\n", appeal, err)
	return &pb.AppealReviewReply{
		AppealID: appeal.AppealID,
	}, nil
}

// AuditAppeal 运营对商家的申诉进行处理,传入参数为AppealID、ReviewID、运营人员信息、操作结果
func (s *ReviewService) AuditAppeal(ctx context.Context, req *pb.AuditAppealRequest) (*pb.AuditAppealReply, error) {
	fmt.Printf("[service] AuditAppeal req:%v\n", req)
	// DTO->PO
	err := s.uc.AuditAppeal(ctx, &biz.AuditAppealParam{
		ReviewID: req.ReviewID,
		AppealID: req.AppealID,
		OpUser:   req.OpUser,
		Status:   req.Status,
	})
	// 这里只需要返回错误
	if err != nil {
		return nil, err
	}
	return &pb.AuditAppealReply{}, nil
}

// ListReviewByUserID 获取用户的所有评价信息,传入参数为UserID、Page页码、Size每页的内容条数
func (s *ReviewService) ListReviewByUserID(ctx context.Context, req *pb.ListReviewByUserIDRequest) (*pb.ListReviewByUserIDReply, error) {
	fmt.Printf("[service] ListReviewByUserID req:%v\n", req)
	reviews, err := s.uc.ListReviewByUserID(ctx, req.GetUserID(), int(req.GetPage()), int(req.GetSize()))
	if err != nil {
		return nil, err
	}
	list := make([]*pb.ReviewInfo, 0, len(reviews))
	for _, v := range reviews {
		list = append(list, &pb.ReviewInfo{
			ReviewID:     v.ReviewID,
			UserID:       v.UserID,
			OrderID:      v.OrderID,
			Score:        v.Score,
			ServiceScore: v.ServiceScore,
			ExpressScore: v.ExpressScore,
			Content:      v.Content,
			PicInfo:      v.PicInfo,
			VideoInfo:    v.VideoInfo,
			Status:       v.Status,
		})
	}
	return &pb.ListReviewByUserIDReply{List: list}, nil
}

// ListReviewByStoreID 获取商户所有被评价信息,传入参数为StoreID、Page页码、Size每页的内容条数
func (s *ReviewService) ListReviewByStoreID(ctx context.Context, req *pb.ListReviewByStoreIDRequest) (*pb.ListReviewByStoreIDReply, error) {
	fmt.Printf("[service] ListReviewByStoreID req:%v\n", req)
	myReviews, err := s.uc.ListReviewByStoreID(ctx, req.GetStoreID(), int(req.GetPage()), int(req.GetSize()))
	if err != nil {
		return nil, err
	}
	// 封装结果返回
	list := make([]*pb.ReviewInfo, 0, len(myReviews))
	for _, v := range myReviews {
		list = append(list, &pb.ReviewInfo{
			ReviewID:     v.ReviewID,
			UserID:       v.UserID,
			OrderID:      v.OrderID,
			Score:        v.Score,
			ServiceScore: v.ServiceScore,
			ExpressScore: v.ExpressScore,
			Content:      v.Content,
			PicInfo:      v.PicInfo,
			VideoInfo:    v.VideoInfo,
			Status:       v.Status,
		})
	}
	return &pb.ListReviewByStoreIDReply{List: list}, nil
}
