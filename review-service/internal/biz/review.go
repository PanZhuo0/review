package biz

import (
	"context"
	"fmt"
	v1 "review-service/api/review/v1"
	"review-service/internal/data/model"
	"review-service/pkg/snowflake"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type ReviewUsecase struct {
	repo ReviewRepo  //biz层需要调用data层的ReviewRepo,同时为了确保biz层提供了对应的方法,biz层会定义一个接口等待data层的repo实现
	log  *log.Helper //从上层链式传递下来的log helper,用来记录可能的跨层信息传递
}

func NewReviewUsecase(repo ReviewRepo, logger log.Logger) *ReviewUsecase {
	return &ReviewUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

type ReviewRepo interface {
	// 这里需要定义biz层要求data层repo需要实现的方法,以SaveReview方法为例
	SaveReview(context.Context, *model.ReviewInfo) (*model.ReviewInfo, error)
	GetReviewByOrderID(context.Context, int64) ([]*model.ReviewInfo, error)
	GetReview(context.Context, int64) (*model.ReviewInfo, error)
	SaveReply(context.Context, *model.ReviewReplyInfo) (*model.ReviewReplyInfo, error)
	GetReviewReply(context.Context, int64) (*model.ReviewReplyInfo, error)
	AuditReview(context.Context, *AuditParam) error
	AppealReview(context.Context, *AppealParam) (*model.ReviewAppealInfo, error)
	AuditAppeal(context.Context, *AuditAppealParam) error
	ListReviewByUserID(ctx context.Context, userID int64, offset, limit int) ([]*model.ReviewInfo, error)
	ListReviewByStoreID(ctx context.Context, storeID int64, offset, limit int) ([]*MyReviewInfo, error)
}

// biz层提供给service层的方法,创建评价方法
// biz层需要实现具体的业务逻辑
func (uc ReviewUsecase) CreateReview(ctx context.Context, review *model.ReviewInfo) (*model.ReviewInfo, error) {
	uc.log.WithContext(ctx).Debugf("[biz] CreateReview,req:%#v", review)
	// 1.参数校验
	// 1.1查看用户是否已经该Order做了评价
	reviews, err := uc.repo.GetReviewByOrderID(ctx, review.OrderID)
	if err != nil {
		return nil, v1.ErrorDbFailed("查询数据库失败")
	}
	if len(reviews) > 0 {
		// 1.2如果用户已经对该Order进行了评价,则直接返回
		return nil, v1.ErrorOrderReviewed("订单号:%d已做过评价", review.OrderID)
	}
	// 2.生成reviewID(使用雪花算法生成)
	reviewID := snowflake.GenID()
	review.ReviewID = reviewID
	// 3.查询订单和商品信息
	// 实际场景中需要通过RPC调用B端查询商品信息,订单具体信息,这里不用实现
	// 4.拼装数据入库
	return uc.repo.SaveReview(ctx, review)
}

/*
GetReview方法 根据评价ID获取评价信息
返回Review信息，需要传入ReviewID
*/
func (uc ReviewUsecase) GetReview(ctx context.Context, reviewID int64) (*model.ReviewInfo, error) {
	uc.log.WithContext(ctx).Debugf("[biz] GetReview reviewID:%v", reviewID)
	return uc.repo.GetReview(ctx, reviewID)
}

/*
CreateReply 创建评价回复
返回一个Reply评价回复对象
*/
func (uc ReviewUsecase) CreateReply(ctx context.Context, param *ReplyParam) (*model.ReviewReplyInfo, error) {
	// 调用data层创建一个用户评价的回复(商家回复用户)
	uc.log.WithContext(ctx).Debugf("[biz] CreateReply param:%v", param)
	// DTO->PO
	reply := model.ReviewReplyInfo{
		ReplyID:   snowflake.GenID(),
		ReviewID:  param.ReviewID,
		StoreID:   param.StoreID,
		Content:   param.Content,
		PicInfo:   param.PicInfo,
		VideoInfo: param.VideoInfo,
	}
	return uc.repo.SaveReply(ctx, &reply)
}

// AuditReview 审核评价
func (uc ReviewUsecase) AuditReview(ctx context.Context, param *AuditParam) error {
	uc.log.WithContext(ctx).Debugf("[biz] AuditReveiw param:%v", param)
	return uc.repo.AuditReview(ctx, param)
}

// AppealReview 申诉评价
func (uc ReviewUsecase) AppealReview(ctx context.Context, param *AppealParam) (*model.ReviewAppealInfo, error) {
	uc.log.WithContext(ctx).Debugf("[biz] AppealReview param:%v", param)
	return uc.repo.AppealReview(ctx, param)
}

// AudiAppeal 审核申诉
func (uc ReviewUsecase) AuditAppeal(ctx context.Context, param *AuditAppealParam) error {
	uc.log.WithContext(ctx).Debugf("[biz] AuditAppeal param:%v", param)
	return uc.repo.AuditAppeal(ctx, param)
}

// ListReviewByID 根据UserID分页查询评价
func (uc ReviewUsecase) ListReviewByUserID(ctx context.Context, userID int64, page, size int) ([]*model.ReviewInfo, error) {
	// 如果页码为负数、默认返回第一页
	if page <= 0 {
		page = 1
	}
	// 如果size不合规、默认使用10页
	if size <= 0 || size > 50 {
		size = 10
	}
	offset := (page - 1) * size
	limit := size
	uc.log.WithContext(ctx).Debugf("[biz] ListReviewByID userID:%v", userID)
	return uc.repo.ListReviewByUserID(ctx, userID, offset, limit)
}

// ListReviewByStoreID 根据store商家ID进行分页评价查询
func (uc ReviewUsecase) ListReviewByStoreID(ctx context.Context, storeID int64, page, size int) ([]*MyReviewInfo, error) {
	// 如果页面不合规，默认页面为1
	if page <= 0 {
		page = 1
	}
	// 如果每页记录条数不合规，默认每页记录条数10条
	if size <= 0 || size > 50 {
		size = 10
	}
	offset := (page - 1) * size
	limit := size
	uc.log.WithContext(ctx).Debugf("[biz] ListReviewStoreID storeID:%v", storeID)
	return uc.repo.ListReviewByStoreID(ctx, storeID, offset, limit)
}

// 解决时间的JSON解析问题
type MyReviewInfo struct {
	*model.ReviewInfo
	CreateAt MyTime `json:"create_at"` // 创建时间,时间类型需要重写UnmarshalJSON方法
	UpdateAt MyTime `json:"update_at"` // 创建时间,时间类型需要重写UnmarshalJSON方法
	// DeleteAt     *time.Time `json:"delete_at"` // 删除时间
	Anonymous    int32 `json:"anonymous,string"` // int32类型的字段需要，string 标记，标记这些数据是从json的string类型反序列化到int32
	Score        int32 `json:"score,string"`
	ServiceScore int32 `json:"service_score,string"`
	ExpressScore int32 `json:"express_score,string"`
	HasMedia     int32 `json:"has_media,string"`
	Status       int32 `json:"status,string"`
	IsDefault    int32 `json:"is_default,string"`
	HasReply     int32 `json:"has_reply,string"`
	ID           int64 `json:"id,string"`
	Version      int32 `json:"version,string"`
	ReviewID     int64 `json:"review_id,string"`
	OrderID      int64 `json:"order_id,string"`
	SkuID        int64 `json:"sku_id,string"`
	SpuID        int64 `json:"spu_id,string"`
	StoreID      int64 `json:"store_id,string"`
	UserID       int64 `json:"user_id,string"`
}

type MyTime time.Time

// MyTime类型的Unmarshal方法,json.unmarshal时会自动调用该方法
func (t *MyTime) UnmarshalJSON(data []byte) error {
	// data = "\"2023-12-17 14:20:18\""
	s := strings.Trim(string(data), `"`)
	fmt.Println("string:", s)
	tmp, err := time.Parse("2006-01-02 15:04:05", s) //以这种形式解析
	if err != nil {
		return err
	}
	*t = MyTime(tmp)
	return nil
}
