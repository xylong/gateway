package controller

import (
	"encoding/json"
	"github.com/e421083458/gateway/dao"
	"github.com/e421083458/gateway/dto"
	"github.com/e421083458/gateway/middleware"
	"github.com/e421083458/gateway/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminLoginController struct {
}

func AdminLoginRegister(group *gin.RouterGroup) {
	adminLogin := &AdminLoginController{}
	group.POST("/login", adminLogin.Login)
	group.GET("/logout", adminLogin.Logout)
}

// Login godoc
// @Summary 登陆
// @Description 管理员登陆
// @Tags 管理员接口
// @ID /admin_login/login
// @Accept json
// @Produce json
// @Param body body dto.AdminLoginInput true "body"
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOutPut} "success"
// @Router /admin_login/login [post]
func (c *AdminLoginController) Login(ctx *gin.Context) {
	input := &dto.AdminLoginInput{}
	if err := input.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 1001, err)
	}
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}
	admin := &dao.Admin{}
	err = admin.LoginCheck(ctx, tx, input)
	if err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}
	// 设置session
	sessionInfo := &dto.AdminSession{
		Id:        admin.Id,
		UserName:  admin.UserName,
		LoginTime: time.Now(),
	}
	sessionByte, err := json.Marshal(sessionInfo)
	if err != nil {
		middleware.ResponseError(ctx, 2003, err)
		return
	}
	session := sessions.Default(ctx)
	session.Set(public.AdminSessionKey, string(sessionByte))
	session.Save()

	middleware.ResponseSuccess(ctx, &dto.AdminLoginOutPut{
		Token: admin.UserName,
	})
}

// Login godoc
// @Summary 登出
// @Description 管理员登出
// @Tags 管理员接口
// @ID /admin_login/logout
// @Accept json
// @Produce json
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin_login/logout [get]
func (c *AdminLoginController) Logout(ctx *gin.Context) {
	session := sessions.Default(ctx)
	session.Delete(public.AdminSessionKey)
	session.Save()

	middleware.ResponseSuccess(ctx, "")
}
