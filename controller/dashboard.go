package controller

import (
	"errors"
	"github.com/e421083458/gateway/dao"
	"github.com/e421083458/gateway/dto"
	"github.com/e421083458/gateway/middleware"
	"github.com/e421083458/gateway/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"time"
)

type DashboardController struct {
}

func DashboardRegister(group *gin.RouterGroup) {
	ctrl := &DashboardController{}
	group.GET("/panel", ctrl.Panel)
	group.GET("/flow_stat", ctrl.FlowStat)
	group.GET("/service_stat", ctrl.ServiceStat)
}

// PanelGroupData godoc
// @Summary 指标统计
// @Description 指标统计
// @Tags 首页大盘
// @ID /dashboard/panel
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.PanelOutput} "success"
// @Router /dashboard/panel [get]
func (c *DashboardController) Panel(ctx *gin.Context) {
	var (
		tx         *gorm.DB
		err        error
		serviceNum int64
		appNum     int64
	)
	if tx, err = lib.GetGormPool("default"); err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	service := &dao.ServiceInfo{}
	if _, serviceNum, err = service.Page(ctx, tx, &dto.ServiceListInput{
		PageNo:   1,
		PageSize: 1,
	}); err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}

	app := &dao.App{}
	if _, appNum, err = app.Select(ctx, tx, &dto.APPListInput{
		PageNo:   1,
		PageSize: 1,
	}); err != nil {
		middleware.ResponseError(ctx, 2003, err)
		return
	}

	out := &dto.PanelOutput{
		ServiceNum:      serviceNum,
		AppNum:          appNum,
		CurrentQPS:      0,
		TodayRequestNum: 0,
	}
	middleware.ResponseSuccess(ctx, out)
}

// FlowStat godoc
// @Summary 服务统计
// @Description 服务统计
// @Tags 首页大盘
// @ID /dashboard/flow_stat
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.ServiceStatOutput} "success"
// @Router /dashboard/flow_stat [get]
func (c *DashboardController) FlowStat(ctx *gin.Context) {
	var (
		todayList     []int64
		yesterdayList []int64
	)

	for i := 0; i < time.Now().Hour(); i++ {
		todayList = append(todayList, 0)
	}
	for i := 0; i < 24; i++ {
		yesterdayList = append(yesterdayList, 0)
	}
	middleware.ResponseSuccess(ctx, &dto.ServiceStatOutput{
		Today:     todayList,
		Yesterday: yesterdayList,
	})
}

// ServiceStat godoc
// @Summary 服务统计
// @Description 服务统计
// @Tags 首页大盘
// @ID /dashboard/service_stat
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.DashServiceStatOutput} "success"
// @Router /dashboard/service_stat [get]
func (c *DashboardController) ServiceStat(ctx *gin.Context) {
	var (
		err  error
		tx   *gorm.DB
		list []dto.DashServiceStatItemOutput
	)
	if tx, err = lib.GetGormPool("default"); err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}
	info := &dao.ServiceInfo{}
	if list, err = info.GroupByLoadType(ctx, tx); err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}
	legend := []string{}
	for index, item := range list {
		name, ok := public.LoadTypeMap[item.LoadType]
		if !ok {
			middleware.ResponseError(ctx, 2003, errors.New("load_type not found"))
			return
		}
		list[index].Name = name
		legend = append(legend, name)
	}

	middleware.ResponseSuccess(ctx, &dto.DashServiceStatOutput{
		Legend: legend,
		Data:   list,
	})
}
