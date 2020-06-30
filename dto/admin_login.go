package dto

import (
	"github.com/e421083458/gateway/public"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminLoginInput struct {
	UserName string `json:"username" form:"username" comment:"姓名" example:"admin" validate:"required,is_valid_username"` // 管理员账号
	Password string `json:"password" form:"password" comment:"密码" example:"123456" validate:"required"`                  // 登陆密码
}

func (input *AdminLoginInput) BindValidParam(ctx *gin.Context) error {
	return public.DefaultGetValidParams(ctx, input)
}

type AdminLoginOutPut struct {
	Token string `json:"token" from:"token" comment:"token" example:"token" validate:""` // 令牌
}

type AdminSession struct {
	Id        int       `json:"id"`
	UserName  string    `json:"username"`
	LoginTime time.Time `json:"login_time"`
}
