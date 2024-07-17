package service

import (
	"context"

	pb "operator/api/operator/v1"
	"operator/internal/biz"
)

type OperationService struct {
	pb.UnimplementedOperationServer
	uc *biz.OperationUsecase
}

func NewOperationService(uc *biz.OperationUsecase) *OperationService {
	return &OperationService{
		uc: uc,
	}
}

// AuditReview 运营端对用户评论进行审核
func (s *OperationService) AuditReview(ctx context.Context, req *pb.AuditReviewRequest) (*pb.AuditReviewReply, error) {
	// DTO->PO
	err := s.uc.AuditReview(ctx, &biz.AuditReviewParam{
		ReviewID:  req.GetReviewID(),
		Status:    int(req.GetStatus()),
		OpReason:  req.GetOpReason(),
		OpRemarks: req.GetOpRemarks(),
		OpUser:    req.GetOpUser(),
	})
	return &pb.AuditReviewReply{}, err
}

// AuditAppeal 运营端对商家的申诉进行审核
func (s *OperationService) AuditAppeal(ctx context.Context, req *pb.AuditAppealRequest) (*pb.AuditAppealReply, error) {
	// DTO->PO
	err := s.uc.AuditAppeal(ctx, &biz.AuditAppealParam{
		AppealID:  req.GetAppealID(),
		ReviewID:  req.GetReviewID(),
		Status:    int(req.GetStatus()), //审核返回的状态
		OpReason:  req.GetOpReason(),
		OpRemarks: req.GetOpRemarks(),
		OpUser:    req.GetOpUser(),
	})
	return &pb.AuditAppealReply{}, err
}
