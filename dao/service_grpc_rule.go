package dao

import (
	"github.com/e421083458/gateway/public"
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
)

type GrpcRule struct {
	ID             int64  `json:"id" gorm:"primary_key"`
	ServiceID      int64  `json:"service_id" gorm:"column:service_id" description:"服务id	"`
	Port           int    `json:"port" gorm:"column:port" description:"端口	"`
	HeaderTransfor string `json:"header_transfor" gorm:"column:header_transfor" description:"header转换支持增加(add)、删除(del)、修改(edit) 格式: add headname headvalue"`
}

func (rule *GrpcRule) TableName() string {
	return "gateway_service_grpc_rule"
}

func (rule *GrpcRule) Find(c *gin.Context, tx *gorm.DB) error {
	return tx.SetCtx(public.GetGinTraceContext(c)).Where(rule).Find(rule).Error
}

func (rule *GrpcRule) Save(c *gin.Context, tx *gorm.DB) error {
	if err := tx.SetCtx(public.GetGinTraceContext(c)).Save(rule).Error; err != nil {
		return err
	}
	return nil
}

func (rule *GrpcRule) ListByServiceID(c *gin.Context, tx *gorm.DB, serviceID int64) ([]GrpcRule, int64, error) {
	var list []GrpcRule
	var count int64
	query := tx.SetCtx(public.GetGinTraceContext(c))
	query = query.Table(rule.TableName()).Select("*")
	query = query.Where("service_id=?", serviceID)
	err := query.Order("id desc").Find(&list).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	errCount := query.Count(&count).Error
	if errCount != nil {
		return nil, 0, err
	}
	return list, count, nil
}
