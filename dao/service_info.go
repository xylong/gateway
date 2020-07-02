package dao

import (
	"github.com/e421083458/gateway/dto"
	"github.com/e421083458/gateway/public"
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"time"
)

type ServiceInfo struct {
	Id          int64     `json:"id" gorm:"primary_key"`
	LoadType    int       `json:"load_type" gorm:"column:load_type" description:"负载类型 0http 1tcp 2grpc"`
	ServiceName string    `json:"service_name" gorm:"column:service_name" description:"服务名称"`
	ServiceDesc string    `json:"service_desc" gorm:"column:service_desc" description:"服务描述"`
	CreatedAt   time.Time `json:"create_at" gorm:"column:create_at" description:"创建时间"`
	UpdatedAt   time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	IsDelete    int       `json:"is_delete" gorm:"column:is_delete" description:"删除"`
}

func (info *ServiceInfo) TableName() string {
	return "gateway_service_info"
}

func (info *ServiceInfo) ServiceDetail(ctx *gin.Context, db *gorm.DB) (detail *ServiceDetail, err error) {
	type dao interface {
		Find(c *gin.Context, tx *gorm.DB) error
	}

	httpRule := &HttpRule{ServiceID: info.Id}
	tcpRule := &TcpRule{ServiceID: info.Id}
	grpcRule := &GrpcRule{ServiceID: info.Id}
	accessControl := &AccessControl{ServiceID: info.Id}
	loadBalance := &LoadBalance{ServiceID: info.Id}

	rules := [5]dao{
		httpRule,
		tcpRule,
		grpcRule,
		accessControl,
		loadBalance,
	}

	for _, rule := range rules {
		if err = rule.Find(ctx, db); err != nil && err != gorm.ErrRecordNotFound {
			continue
		}
	}

	detail = &ServiceDetail{
		Info:          info,
		HTTPRule:      httpRule,
		TCPRule:       tcpRule,
		GRPCRule:      grpcRule,
		LoadBalance:   loadBalance,
		AccessControl: accessControl,
	}
	return
}

func (info *ServiceInfo) Page(ctx *gin.Context, db *gorm.DB, input *dto.ServiceListInput) (list []ServiceInfo, total int64, err error) {
	offset := (input.PageNo - 1) * input.PageSize
	query := db.SetCtx(public.GetGinTraceContext(ctx))
	query = query.Table(info.TableName()).Where("is_delete=0")
	if input.Info != "" {
		query = query.Where("(service_name like ? or service_desc like ?)", "%"+input.Info+"%", "%"+input.Info+"%")
	}
	if err = query.Limit(input.PageSize).Offset(offset).Order("id desc").Find(&list).Error; err != nil && err != gorm.ErrRecordNotFound {
		return
	}
	query.Limit(input.PageSize).Offset(offset).Count(&total)
	return
}

// Find 查找
func (info *ServiceInfo) Find(ctx *gin.Context, db *gorm.DB) (err error) {
	return db.SetCtx(public.GetGinTraceContext(ctx)).Where(info).Find(info).Error
}

// Save 保存
func (info *ServiceInfo) Save(ctx *gin.Context, db *gorm.DB) error {
	return db.SetCtx(public.GetGinTraceContext(ctx)).Save(info).Error
}
