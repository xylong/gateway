package controller

import (
	"errors"
	"github.com/e421083458/gateway/dao"
	"github.com/e421083458/gateway/dto"
	"github.com/e421083458/gateway/middleware"
	"github.com/e421083458/gateway/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
)

type AppController struct {
}

func AppRegister(group *gin.RouterGroup) {
	app := &AppController{}
	group.GET("apps", app.Index)
	group.GET("app", app.Show)
	group.POST("apps", app.Add)
	group.PUT("apps", app.Update)
	group.DELETE("apps", app.Delete)
	group.GET("statistics", app.Statistics)
}

// APPList godoc
// @Summary 租户列表
// @Description 租户列表
// @Tags 租户管理
// @ID /app/index
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page_size query string true "每页多少条"
// @Param page_no query string true "页码"
// @Success 200 {object} middleware.Response{data=dto.APPListOutput} "success"
// @Router /apps [get]
func (c *AppController) Index(ctx *gin.Context) {
	var (
		list       []dao.App
		total      int64
		err        error
		outputList []dto.APPListItemOutput
	)
	params := &dto.APPListInput{}
	if err = params.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}
	app := &dao.App{}
	if list, total, err = app.Select(ctx, lib.GORMDefaultPool, params); err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}
	for _, item := range list {
		outputList = append(outputList, dto.APPListItemOutput{
			ID:       item.ID,
			AppID:    item.AppID,
			Name:     item.Name,
			Secret:   item.Secret,
			WhiteIPS: item.WhiteIPS,
			Qpd:      item.Qpd,
			RealQpd:  0,
			Qps:      item.Qps,
			RealQps:  0,
		})
	}
	middleware.ResponseSuccess(ctx, &dto.APPListOutput{
		List:  outputList,
		Total: total,
	})
}

// APPDetail godoc
// @Summary 租户详情
// @Description 租户详情
// @Tags 租户管理
// @ID /app/show
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=dao.App} "success"
// @Router /app [get]
func (c *AppController) Show(ctx *gin.Context) {
	var err error
	params := &dto.APPDetailInput{}
	if err = params.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}
	app := &dao.App{
		ID: params.ID,
	}
	if err = app.Find(ctx, lib.GORMDefaultPool); err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}
	middleware.ResponseSuccess(ctx, app)
}

// AppAdd godoc
// @Summary 租户添加
// @Description 租户添加
// @Tags 租户管理
// @ID /app/add
// @Accept  json
// @Produce  json
// @Param body body dto.APPAddHttpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /apps [post]
func (c *AppController) Add(ctx *gin.Context) {
	params := &dto.APPAddHttpInput{}
	if err := params.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}
	//验证app_id是否被占用
	app := &dao.App{
		AppID: params.AppID,
	}
	if err := app.Find(ctx, lib.GORMDefaultPool); err == nil {
		middleware.ResponseError(ctx, 2002, errors.New("租户ID被占用，请重新输入"))
		return
	}
	if params.Secret == "" {
		params.Secret = public.MD5(params.AppID)
	}
	tx := lib.GORMDefaultPool
	app.Name = params.Name
	app.Secret = params.Secret
	app.WhiteIPS = params.WhiteIPS
	app.Qps = params.Qps
	app.Qpd = params.Qpd
	if err := app.Save(ctx, tx); err != nil {
		middleware.ResponseError(ctx, 2003, err)
		return
	}
	middleware.ResponseSuccess(ctx, "")
}

// AppUpdate godoc
// @Summary 租户更新
// @Description 租户更新
// @Tags 租户管理
// @ID /app/update
// @Accept  json
// @Produce  json
// @Param body body dto.APPUpdateHttpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /apps [put]
func (c *AppController) Update(ctx *gin.Context) {
	params := &dto.APPUpdateHttpInput{}
	if err := params.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}
	app := &dao.App{
		ID: params.ID,
	}
	if err := app.Find(ctx, lib.GORMDefaultPool); err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}
	if params.Secret == "" {
		params.Secret = public.MD5(params.AppID)
	}
	app.Name = params.Name
	app.Secret = params.Secret
	app.WhiteIPS = params.WhiteIPS
	app.Qps = params.Qps
	app.Qpd = params.Qpd
	if err := app.Save(ctx, lib.GORMDefaultPool); err != nil {
		middleware.ResponseError(ctx, 2003, err)
		return
	}
	middleware.ResponseSuccess(ctx, "")
}

// APPDelete godoc
// @Summary 租户删除
// @Description 租户删除
// @Tags 租户管理
// @ID /app/delete
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /apps [delete]
func (c *AppController) Delete(ctx *gin.Context) {
	params := &dto.APPDetailInput{}
	if err := params.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	app := &dao.App{
		ID: params.ID,
	}
	if err := app.Find(ctx, lib.GORMDefaultPool); err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}
	app.IsDelete = 1
	if err := app.Save(ctx, lib.GORMDefaultPool); err != nil {
		middleware.ResponseError(ctx, 2003, err)
		return
	}
	middleware.ResponseSuccess(ctx, "")
	return
}

func (c *AppController) Statistics(ctx *gin.Context) {

}
