package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"review-service/internal/biz"
	"review-service/internal/data/model"
	"review-service/internal/data/query"
	"review-service/pkg/snowflake"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type reviewRepo struct {
	data *Data //这里使用的是GORM-GEN生成的query.Query对象作为数据处理结构体
	log  *log.Helper
}

func NewReviewRepo(data *Data, logger log.Logger) biz.ReviewRepo {
	return reviewRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// SaveReview 将Review保存到数据库中 -data层
// 需要传入一个Review对象,返回传入的review对象，以及可能的错误
func (r reviewRepo) SaveReview(ctx context.Context, review *model.ReviewInfo) (ret *model.ReviewInfo, err error) {
	err = r.data.query.ReviewInfo.
		WithContext(ctx).
		Save(review)
	return review, err
}

// GetReviewByOrderID 通过OrderID获取Review信息 data层
// 需要传入一个OrderID,返回Review对象，以及可能的错误
// 根据订单ID获取Review评价信息,需要传入一个OrderID
func (r reviewRepo) GetReviewByOrderID(ctx context.Context, orderID int64) ([]*model.ReviewInfo, error) {
	return r.data.query.
		WithContext(ctx).ReviewInfo.
		Where(r.data.query.ReviewInfo.OrderID.Eq(orderID)).
		Find()
}

// GetReview 获取Review信息,通过ReviewID获取Review信息
// 需要传入一个ReviewID，返回Review对象，以及可能的错误
func (r reviewRepo) GetReview(ctx context.Context, reviewID int64) (*model.ReviewInfo, error) {
	return r.data.query.ReviewInfo.
		WithContext(ctx).
		Where(r.data.query.ReviewInfo.ReviewID.Eq(reviewID)).
		First()
}

// SaveReply 保存商家的回复信息,
// 需要传入一个ReviewReplyInfo对象,返回传入的对象，以及可能的错误
func (r reviewRepo) SaveReply(ctx context.Context, reply *model.ReviewReplyInfo) (*model.ReviewReplyInfo, error) {
	// 1.数据校验，确保商家是否可以回复
	// 1.1 数据合法性校验(商家已回复则不应该重复回复)
	review, err := r.data.query.ReviewInfo.
		WithContext(ctx).
		Where(r.data.query.ReviewInfo.ReviewID.Eq(reply.ReviewID)).
		First()
	if err != nil {
		// 查询过程中出错，返回错误
		return nil, err
	}
	if review.HasReply == 1 {
		return nil, errors.New("该评价商家已进行过回复")
	}
	// 1.2 水平越权校验(A商家只应该能回复自家客户的评论，而不能回复B商家的)
	if review.StoreID != reply.StoreID {
		// 当前商店不是用户评价的那家店,也就是出现水平越权
		return nil, errors.New("水平越权")
	}
	// 2. 通过校验,更新数据库中的数据,保存这条Reply
	// 2.1涉及事务(将Reply插入到ReviewReply表,同时将这个Reveiw的HasReply设置为1,即已回复过)
	r.data.query.Transaction(func(tx *query.Query) error {
		// 将Reply插入ReviewReply表
		if err := tx.ReviewReplyInfo.
			WithContext(ctx).
			Save(reply); err != nil {
			r.log.WithContext(ctx).Errorf("SaveReply create reply failed,err:%v", err)
			return err
		}
		// 更新Review表对应ReviewID记录的HasReply字段
		if _, err := tx.ReviewInfo.WithContext(ctx).Where(tx.ReviewInfo.ReviewID.Eq(reply.ReviewID)).Update(tx.ReviewInfo.HasReply, 1); err != nil {
			r.log.Errorf("SaveReply update review failed,err:%v", err)
			return err
		}
		return nil
	})
	// 返回
	return reply, nil
}

// GetReviewReply 获取商家的回复信息
// 需要传入一个ReviewID，返回ReviewReplyInfo对象，以及可能的错误
func (r reviewRepo) GetReviewReply(ctx context.Context, reviewID int64) (*model.ReviewReplyInfo, error) {
	return r.data.query.ReviewReplyInfo.
		WithContext(ctx).
		Where(r.data.query.ReviewReplyInfo.ReviewID.Eq(reviewID)).
		First()
}

// AuditReview运营对用户评价进行审核
// 需要传入一个审核参数对象AuditParam,返回可能的错误
func (r reviewRepo) AuditReview(ctx context.Context, param *biz.AuditParam) error {
	// 更新用户的评价
	_, err := r.data.query.ReviewInfo.
		WithContext(ctx).
		Where(r.data.query.ReviewInfo.ReviewID.Eq(param.ReviewID)).
		Updates(map[string]interface{}{
			"status":     param.Status,
			"op_user":    param.OpUser,
			"op_reason":  param.OpReason,
			"op_remarks": param.OpRemarks,
		})
	return err
}

// AppealReview 商家对用户的评价进行申诉
// 需要传入一个申诉参数对象AppealParam,返回申诉记录对象,以及可能的错误
func (r reviewRepo) AppealReview(ctx context.Context, param *biz.AppealParam) (*model.ReviewAppealInfo, error) {
	// 1.判断传入的ReviewID记录是否存在
	review, err := r.data.query.ReviewInfo.WithContext(ctx).Where(query.ReviewInfo.ReviewID.Eq(param.ReviewID)).First()
	fmt.Println(review)
	if err != nil {
		return nil, fmt.Errorf("ReviewID:%v,这条评论不存在", param.ReviewID)
	}
	// 2.判断商家是否有权限对该评论申诉
	if review.StoreID != param.StoreID {
		return nil, fmt.Errorf("水平越权,ReviewID:%v这条评论不是对该商家的评论", param.ReviewID)
	}
	// 3.判断申诉记录是否已存在,如果存在，并且该申诉已被处理(status > 10)，就返回
	appeal, err := r.data.query.ReviewAppealInfo.WithContext(ctx).Where(
		query.ReviewAppealInfo.ReviewID.Eq(param.ReviewID),
		query.ReviewAppealInfo.StoreID.Eq(param.StoreID),
	).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// 如果查询出错了，并且不是ErrRecordNotFound错误,则直接返回
		return nil, err
	}
	newAppeal := &model.ReviewAppealInfo{
		ReviewID:  param.ReviewID,
		StoreID:   param.StoreID,
		Status:    10, //Attention:注意这里要被设置为10(未审核状态)，因为记录需要等待运营端的审核
		Reason:    param.Reason,
		Content:   param.Content,
		PicInfo:   param.PicInfo,
		VideoInfo: param.VideoInfo,
	}
	// 如果申诉记录存在
	if appeal != nil {
		if appeal.Status > 10 {
			// 3.1 如果申诉存在且已经审核过,则直接返回
			return nil, fmt.Errorf("AppealID:%v这条申诉已经审核过,请勿重复提交,如果对结果不满意,请致电:110", appeal.AppealID)
		}
		// 如果记录存在,但还未审核，则商家对申诉信息进行更新
		newAppeal.AppealID = appeal.AppealID
		r.data.query.ReviewAppealInfo.
			WithContext(ctx).
			Where(r.data.query.ReviewAppealInfo.AppealID.Eq(appeal.AppealID)).
			Updates(map[string]interface{}{
				"status":     newAppeal.Status,
				"content":    newAppeal.Content,
				"reason":     newAppeal.Reason,
				"pic_info":   newAppeal.PicInfo,
				"video_info": newAppeal.VideoInfo,
			})
	} else {
		// 如果Appeal申诉记录不存在,则设置ID,并将这条申诉请求入库
		newAppeal.AppealID = snowflake.GenID()
		err = r.data.query.ReviewAppealInfo.WithContext(ctx).Save(newAppeal)
		r.log.Errorf("将Appeal申诉记录存入数据库时出错,Appeal:%v", newAppeal)
		if err != nil {
			return nil, err
		}
		return newAppeal, nil
	}

	return newAppeal, err
}

// AppealReview operator运营对商家的申诉进行处理
func (r reviewRepo) AuditAppeal(ctx context.Context, param *biz.AuditAppealParam) error {
	// 1.判断请求是否合法
	fmt.Println(param)
	// 1.1 appeal记录是否存在
	_, err := r.data.query.ReviewAppealInfo.WithContext(ctx).Where(r.data.query.ReviewAppealInfo.AppealID.Eq(param.AppealID)).First()
	if err != nil {
		return fmt.Errorf("获取数据时出现错误:%v", err)
	}
	// 使用事务，1.更新申诉表，2.如果审核中的结果是20（即通过），还需要把用户评价的Status设置为40(隐藏评价)
	err = r.data.query.Transaction(func(tx *query.Query) error {
		// 1.申诉表的更新
		if _, err := tx.ReviewAppealInfo.WithContext(ctx).Where(tx.ReviewAppealInfo.AppealID.Eq(param.AppealID)).Updates(
			map[string]interface{}{
				"status":  param.Status,
				"op_user": param.OpUser,
				// op_reason:param居然没设置op_reason
			}); err != nil {
			return err
		}
		// 2.如果审核通过
		if param.Status == 20 {
			if _, err := tx.ReviewInfo.
				WithContext(ctx).
				Where(tx.ReviewInfo.ReviewID.Eq(param.ReviewID)).
				Update(tx.ReviewInfo.Status, 40); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// ListReviewByUserID 列举出用户的所有评价
func (r reviewRepo) ListReviewByUserID(ctx context.Context, userID int64, offset, limit int) ([]*model.ReviewInfo, error) {
	return r.data.query.ReviewInfo.
		WithContext(ctx).
		Where(r.data.query.ReviewInfo.UserID.Eq(userID)).
		Offset(offset).
		Limit(limit).
		Find()
}

func (r *reviewRepo) getData1(ctx context.Context, storeID int64, offset, limit int) ([]*biz.MyReviewInfo, error) {
	// 去ES中查询Review
	resp, err := r.data.es.
		Search().
		Index("review").
		From(offset).
		Size(limit).
		Query(&types.Query{
			Bool: &types.BoolQuery{
				Filter: []types.Query{
					{
						Term: map[string]types.TermQuery{
							"store_id": {Value: storeID},
						},
					},
				},
			},
		}).Do(ctx)
	if err != nil {
		r.log.Debugf("使用ES查询失败,err:", err)
		return nil, err
	}
	fmt.Printf("es result numbers total :%v\n", resp.Hits.Total.Value)
	// 把从ES中获取的数据反序列化为MyReviewInfo形式
	list := make([]*biz.MyReviewInfo, 0, resp.Hits.Total.Value)
	for _, hit := range resp.Hits.Hits {
		tmp := &biz.MyReviewInfo{}
		if err := json.Unmarshal(hit.Source_, tmp); err != nil {
			/*
				err:parsing time "2024-07-19 10:41:52" as "2006-01-02T15:04:05Z07:00": cannot parse " 10:41:52" as "T"
				将ES中的数据反序列化到time.Time出错
				ES中的时间格式：2024-07-19 10:30:11	不能解析为time.Time
				go中的时间格式: 2006-01-02T15:04:05Z07:00
			*/
			r.log.Errorf("json.Unmarshal(hit.Source_,tmp),err:%v", err)
			continue
		}
		list = append(list, tmp)
	}
	return list, nil
}

// ListReviewByStoreID 列举所有对商户的评价
func (r reviewRepo) ListReviewByStoreID(ctx context.Context, storeID int64, offset, limit int) ([]*biz.MyReviewInfo, error) {
	// return r.getData1(ctx, storeID, offset, limit) //直接查ES
	return r.getData2(ctx, storeID, offset, limit) //增加缓存和Single flight
}

var g singleflight.Group

// KEY的设计:review:store_id:offset:size
func (r *reviewRepo) getData2(ctx context.Context, storeID int64, offset, limit int) ([]*biz.MyReviewInfo, error) {
	// 1.先查询redis缓存
	// 2.缓存没有则查询ES
	// 3.通过single fight合并短时间内大量的并发查询
	key := fmt.Sprintf("review:%d:%d:%d", storeID, offset, limit)
	b, err := r.getDataBySingleflight(ctx, key)
	if err != nil {
		return nil, err
	}
	// 反序列化
	hm := new(types.HitsMetadata)
	if err := json.Unmarshal(b, hm); err != nil {
		return nil, err
	}
	list := make([]*biz.MyReviewInfo, 0, hm.Total.Value)
	for _, hit := range hm.Hits {
		tmp := &biz.MyReviewInfo{}
		if err := json.Unmarshal(hit.Source_, tmp); err != nil {
			r.log.Errorf("json.Unmarshal(hit.Source_, tmp) failed,err:", err)
			continue
		}
		list = append(list, tmp)
	}
	return list, nil
}

func (r *reviewRepo) getDataBySingleflight(ctx context.Context, key string) (data []byte, err error) {
	v, err, sharded := g.Do(key, func() (any, error) {
		// 1.查缓存
		data, err := r.getDataFromCache(ctx, key)
		r.log.Debugf("r.getDataFromCache(ctx, key) data:%v,err:%v", data, err)
		if err == nil {
			return data, nil
		}
		// 2.缓存中没有记录
		if errors.Is(err, redis.Nil) {
			fmt.Println("CACHE")
			fmt.Println("CACHE")
			fmt.Println("CACHE")
			fmt.Println("CACHE")
			fmt.Println("CACHE")
			fmt.Println("CACHE")
			fmt.Println("CACHE")
			fmt.Println("CACHE")
			fmt.Println("CACHE")
			// 缓存中没有这个key,则去查询ES
			data, err := r.getDataFromES(ctx, key)
			if err == nil {
				// (同时将ES数据写入Redis缓存)
				return data, r.writeCache(ctx, key, data)
			}
			// 从ES获取数据时出现了错误
			return nil, err
		}
		// 3.查缓存失败(RedisDone掉了)(直接返回错误，不继续向下传递错误,熔断)
		return nil, err
	})
	r.log.Debugf("singleflight ret: v:%v err:%v shared:%v\n", v, err, sharded)
	if err != nil {
		return nil, err
	}
	return v.([]byte), nil
}

// 读取缓存
func (r *reviewRepo) getDataFromCache(ctx context.Context, key string) ([]byte, error) {
	r.log.Debugf("getDataFromCache key:%v\n", key)
	fmt.Println(r.data.rdb)
	return r.data.rdb.Get(key).Bytes()
}

// 将ES中的数据写入缓存
func (r *reviewRepo) writeCache(ctx context.Context, key string, data []byte) error {
	return r.data.rdb.Set(key, data, time.Minute).Err() //计入缓存，时延1Min
}

// 从ES中获取数据
func (r *reviewRepo) getDataFromES(ctx context.Context, key string) ([]byte, error) {
	values := strings.Split(key, ":")
	if len(values) < 4 {
		// review:store_id:pagenum:size
		return nil, errors.New("invalid key")
	}
	index, storeID, offsetStr, limitStr := values[0], values[1], values[2], values[3]
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return nil, err
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return nil, err
	}
	resp, err := r.data.es.Search().
		Index(index).
		From(offset).
		Size(limit).
		Query(&types.Query{
			Bool: &types.BoolQuery{
				Filter: []types.Query{
					{
						Term: map[string]types.TermQuery{
							"store_id": {Value: storeID},
						},
					},
				},
			},
		}).Do(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(resp.Hits) //返回的是ES中search操作返回的HIT metadata(Hit元数据+hit到的数据)
}
