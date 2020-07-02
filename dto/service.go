package dto

import (
	"github.com/e421083458/gateway/public"
	"github.com/gin-gonic/gin"
)

type ServiceListInput struct {
	Info     string `json:"info" form:"info" comment:"关键词" example:"" validate:""`                      //关键词
	PageNo   int    `json:"page_no" form:"page_no" comment:"页码" example:"1" validate:"required"`        // 页数
	PageSize int    `json:"page_size" form:"page_size" comment:"每页条数" example:"20" validate:"required"` // 每页条数
}

func (input *ServiceListInput) BindValidParam(ctx *gin.Context) error {
	return public.DefaultGetValidParams(ctx, input)
}

type ServiceItemOutput struct {
	Id          int64  `json:"id" form:"id"`                     // id
	ServiceName string `json:"service_name" form:"service_name"` // 服务名称
	ServiceDesc string `json:"service_desc" form:"service_desc"` // 服务描述
	LoadType    int    `json:"load_type" form:"load_type"`       // 类型
	ServiceAddr string `json:"service_addr" form:"service_addr"` // 服务地址
	Qps         int64  `json:"qps" form:"qps"`                   // qps
	Qpd         int64  `json:"qpd" form:"qpd"`                   // qpd
	TotalNode   int    `json:"total_node" form:"total_node"`     // 节点数
}

type ServiceListOutput struct {
	Total int64               `json:"total" form:"total" comment:"总数"` // 总数
	List  []ServiceItemOutput `json:"list" form:"list" comment:"列表"`   // 列表
}

// ServiceDeleteInput 删除服务
type ServiceDeleteInput struct {
	ID int64 `json:"id" form:"id" comment:"服务ID" example:"1" validate:"required"` // 服务ID
}

func (input *ServiceDeleteInput) BindValidParam(ctx *gin.Context) error {
	return public.DefaultGetValidParams(ctx, input)
}
