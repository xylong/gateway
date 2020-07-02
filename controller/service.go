package controller

import (
	"github.com/e421083458/gateway/dao"
	"github.com/e421083458/gateway/dto"
	"github.com/e421083458/gateway/middleware"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
)

type ServiceController struct {
}

func ServiceRegister(group *gin.RouterGroup) {
	service := &ServiceController{}
	group.GET("/index", service.Index)
}

// ServiceList godoc
// @Summary 服务列表
// @Description 服务列表
// @Tags 服务管理
// @ID /service/index
// @Accept json
// @Produce json
// @Param info query string false "关键词"
// @Param page_no query int true "页码"
// @Param page_size query int true "每页条数"
// @Success 200 {object} middleware.Response{data=dto.ServiceListOutput} "success"
// @Router /service/index [get]
func (c *ServiceController) Index(ctx *gin.Context) {
	input := &dto.ServiceListInput{}
	if err := input.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	serviceInfo := &dao.ServiceInfo{}
	list, total, err := serviceInfo.Page(ctx, tx, input)
	if err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}

	outList := []dto.ServiceItemOutput{}
	for _, item := range list {
		outItem := dto.ServiceItemOutput{
			Id:          item.Id,
			ServiceName: item.ServiceName,
			ServiceDesc: item.ServiceDesc,
		}
		outList = append(outList, outItem)
	}

	middleware.ResponseSuccess(ctx, &dto.ServiceListOutput{
		Total: total,
		List:  outList,
	})
}
