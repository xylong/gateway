package dao

import (
	"github.com/e421083458/gateway/public"
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
)

type TcpRule struct {
	ID        int64 `json:"id" gorm:"primary_key"`
	ServiceID int64 `json:"service_id" gorm:"column:service_id" description:"服务id	"`
	Port      int   `json:"port" gorm:"column:port" description:"端口	"`
}

func (rule *TcpRule) TableName() string {
	return "gateway_service_tcp_rule"
}

func (rule *TcpRule) Find(c *gin.Context, tx *gorm.DB) error {
	return tx.SetCtx(public.GetGinTraceContext(c)).Where(rule).Find(rule).Error
}

func (rule *TcpRule) Save(c *gin.Context, tx *gorm.DB) error {
	if err := tx.SetCtx(public.GetGinTraceContext(c)).Save(rule).Error; err != nil {
		return err
	}
	return nil
}

func (rule *TcpRule) ListByServiceID(c *gin.Context, tx *gorm.DB, serviceID int64) ([]TcpRule, int64, error) {
	var list []TcpRule
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
