package dto

import (
	"github.com/e421083458/gateway/public"
	"github.com/gin-gonic/gin"
	"time"
)

// AdminInfoOutput 登录信息
type AdminInfoOutput struct {
	Id           int       `json:"id"`
	Name         string    `json:"name"`
	Avatar       string    `json:"avatar"`
	Introduction string    `json:"introduction"`
	Roles        []string  `json:"roles"`
	LoginTime    time.Time `json:"login_time"`
}

// ChangePasswordInput 新密码
type ChangePasswordInput struct {
	Password string `json:"password" form:"password" comment:"密码" example:"123456" validate:"required"` // 新密码
}

func (input *ChangePasswordInput) BindValidParam(ctx *gin.Context) error {
	return public.DefaultGetValidParams(ctx, input)
}
