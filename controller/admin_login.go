package controller

import (
	"github.com/e421083458/gateway/dto"
	"github.com/e421083458/gateway/middleware"
	"github.com/gin-gonic/gin"
)

type AdminLoginController struct {
	
}

func AdminLoginRegister(group *gin.RouterGroup)  {
	adminLogin:=&AdminLoginController{}
	group.POST("/login",adminLogin.Login)
}

func (c *AdminLoginController) Login(ctx *gin.Context) {
	params:=&dto.AdminLoginInput{}
	if err:=params.BindValidParam(ctx);err!=nil {
		middleware.ResponseError(ctx,1001,err)
	}
	middleware.ResponseSuccess(ctx,"")
}