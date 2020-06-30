package controller

import (
	"encoding/json"
	"fmt"
	"github.com/e421083458/gateway/dto"
	"github.com/e421083458/gateway/middleware"
	"github.com/e421083458/gateway/public"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AdminController struct {
}

func AdminRegister(group *gin.RouterGroup) {
	admin := &AdminController{}
	group.GET("/admin_info", admin.AdminLogin)
}

// Login godoc
// @Summary 管理员信息
// @Description 管理员信息
// @Tags 管理员接口
// @ID /admin/admin_info
// @Accept json
// @Produce json
// @Success 200 {object} middleware.Response{data=dto.AdminInfoOutput} "success"
// @Router /admin/admin_info [get]
func (c *AdminController) AdminLogin(ctx *gin.Context) {
	session := sessions.Default(ctx)
	info := session.Get(public.AdminSessionKey)
	adminSessionInfo := &dto.AdminSession{}
	if err := json.Unmarshal([]byte(fmt.Sprint(info)), adminSessionInfo); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}
	middleware.ResponseSuccess(ctx, &dto.AdminInfoOutput{
		Id:           adminSessionInfo.Id,
		Name:         adminSessionInfo.UserName,
		Avatar:       "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif",
		Introduction: "非吾小天下，才高而已",
		Roles:        []string{"admin"},
		LoginTime:    adminSessionInfo.LoginTime,
	})
}
