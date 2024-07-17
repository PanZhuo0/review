package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// AuditReviewParam 审核评价的参数
type AuditReviewParam struct {
	ReviewID  int64
	Status    int
	OpReason  string
	OpRemarks string
	OpUser    string
}

// AuditAppealParam 审核申诉的参数
type AuditAppealParam struct {
	AppealID  int64
	ReviewID  int64
	StoreID   int64
	Status    int
	OpReason  string
	OpRemarks string
	OpUser    string
}

type OperationUsecase struct {
	log  log.Helper
	repo OpeartionRepo
}

type OpeartionRepo interface {
	AuditReview(context.Context, *AuditReviewParam) error
	AuditAppeal(context.Context, *AuditAppealParam) error
}

func NewOperationUsecase(repo OpeartionRepo, logger log.Logger) *OperationUsecase {
	return &OperationUsecase{
		log:  *log.NewHelper(logger),
		repo: repo,
	}
}

// 审核用户评价评价
func (uc *OperationUsecase) AuditReview(ctx context.Context, param *AuditReviewParam) error {
	uc.log.WithContext(ctx).Infof("AuditReview,param:%v", param)
	return uc.repo.AuditReview(ctx, param)
}

// 审核商家的申诉
func (uc *OperationUsecase) AuditAppeal(ctx context.Context, param *AuditAppealParam) error {
	uc.log.WithContext(ctx).Infof("AuditAppeal,param:%v", param)
	return uc.repo.AuditAppeal(ctx, param)
}
