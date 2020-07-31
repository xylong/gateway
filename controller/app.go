package controller

import (
	"github.com/e421083458/gateway/dao"
	"github.com/e421083458/gateway/dto"
	"github.com/e421083458/gateway/middleware"
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

func (c *AppController) Show(ctx *gin.Context) {

}

func (c *AppController) Add(ctx *gin.Context) {

}

func (c *AppController) Update(ctx *gin.Context) {

}

func (c *AppController) Delete(ctx *gin.Context) {

}

func (c *AppController) Statistics(ctx *gin.Context) {

}
