package data

import (
	"context"
	"fmt"
	v1 "operator/api/review/v1"
	"operator/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type operationRepo struct {
	data *Data
	log  *log.Helper
}

func NewOperationRepo(data *Data, logger log.Logger) biz.OpeartionRepo {
	return &operationRepo{
		log:  log.NewHelper(logger),
		data: data,
	}
}

func (r *operationRepo) AuditReview(ctx context.Context, param *biz.AuditReviewParam) error {
	r.log.WithContext(ctx).Infof("AuditReview, param:%v", param)
	ret, err := r.data.rc.AuditReview(ctx, &v1.AuditReviewRequest{
		ReviewID:  param.ReviewID,
		Status:    int32(param.Status),
		OpUser:    param.OpUser,
		OpReason:  param.OpReason,
		OpRemarks: &param.OpRemarks,
	})
	r.log.WithContext(ctx).Debugf("AuditReview reply ret: %v, err:%v", ret, err)
	return err
}

func (r *operationRepo) AuditAppeal(ctx context.Context, param *biz.AuditAppealParam) error {
	r.log.WithContext(ctx).Infof("AuditReview, param:%v", param)
	fmt.Println(ctx.Err())
	ret, err := r.data.rc.AuditAppeal(ctx, &v1.AuditAppealRequest{
		AppealID:  param.AppealID,
		ReviewID:  param.ReviewID,
		Status:    int32(param.Status),
		OpUser:    param.OpUser,
		OpRemarks: &param.OpRemarks,
	})
	fmt.Println(ctx.Err())
	fmt.Println("-------------------------------- OPERATE,remote RPC:", r.data.rc)
	r.log.WithContext(ctx).Debugf("AuditReview reply ret: %v, err:%v", ret, err)
	return err
}
