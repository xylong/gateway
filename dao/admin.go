package dao

import (
	"github.com/e421083458/gateway/dto"
	"github.com/e421083458/gateway/public"
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"time"
)

type Admin struct {
	Id       int       `json:"id" gorm:"primary_key" description:"自增主键"`
	UserName string    `json:"user_name" gorm:"column:user_name" description:"管理员账户"`
	Salt     string    `json:"salt" gorm:"column:salt" description:"盐"`
	Password string    `json:"password" gorm:"column:password" description:"密码"`
	CreateAt time.Time `json:"create_at" gorm:"column:create_at" description:"创建时间"`
	UpdateAT time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	IsDelete int       `json:"is_delete" gorm:"column:is_delete" description:"删除"`
}

func (admin *Admin) TableName() string {
	return "gateway_admin"
}

func (admin *Admin) Find(ctx *gin.Context, db *gorm.DB) (*Admin, error) {
	err := db.SetCtx(public.GetGinTraceContext(ctx)).Where(admin).Find(admin).Error
	if err != nil {
		return nil, err
	}
	return admin, nil
}

// LoginCheck 登陆判断
func (admin *Admin) LoginCheck(ctx *gin.Context, db *gorm.DB, input *dto.AdminLoginInput) (*Admin, error) {
	admin.UserName = input.UserName
	admin.IsDelete = 0
	info, err := admin.Find(ctx, db)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	saltPassword := public.SaltPassword(info.Salt, input.Password)
	if info.Password != saltPassword {
		return nil, errors.New("密码错误")
	}
	return info, nil
}
