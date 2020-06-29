package controller

import (
	"github.com/e421083458/gateway/dto"
	"github.com/e421083458/gateway/middleware"
	"github.com/gin-gonic/gin"
)

type AdminLoginController struct {
}

func AdminLoginRegister(group *gin.RouterGroup) {
	adminLogin := &AdminLoginController{}
	group.POST("/login", adminLogin.Login)
}

// Login godoc
// @Summary 管理员登陆
// @Description 管理员登陆
// @Tags 管理员接口
// @ID /admin_login/login
// @Accept json
// @Produce json
// @Param body body dto.AdminLoginInput true "body"
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOutPut} "success"
// @Router /admin_login/login [post]
func (c *AdminLoginController) Login(ctx *gin.Context) {
	params := &dto.AdminLoginInput{}
	if err := params.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 1001, err)
	}
	out := &dto.AdminLoginOutPut{
		Token: params.UserName,
	}
	middleware.ResponseSuccess(ctx, out)
}
