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
	Id        int       `json:"id" gorm:"primary_key" description:"自增主键"`
	UserName  string    `json:"user_name" gorm:"column:user_name" description:"管理员账户"`
	Salt      string    `json:"salt" gorm:"column:salt" description:"盐"`
	Password  string    `json:"password" gorm:"column:password" description:"密码"`
	CreatedAt time.Time `json:"create_at" gorm:"column:create_at" description:"创建时间"`
	UpdatedAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	IsDelete  int       `json:"is_delete" gorm:"column:is_delete" description:"删除"`
}

func (admin *Admin) TableName() string {
	return "gateway_admin"
}

// Find 查找用户
func (admin *Admin) Find(ctx *gin.Context, db *gorm.DB) (err error) {
	err = db.SetCtx(public.GetGinTraceContext(ctx)).Where(admin).Find(admin).Error
	return
}

// Save 保存
func (admin *Admin) Save(ctx *gin.Context, db *gorm.DB) error {
	return db.SetCtx(public.GetGinTraceContext(ctx)).Save(admin).Error
}

// LoginCheck 登陆判断
func (admin *Admin) LoginCheck(ctx *gin.Context, db *gorm.DB, input *dto.AdminLoginInput) error {
	admin.UserName = input.UserName
	admin.IsDelete = 0
	err := admin.Find(ctx, db)
	if err != nil {
		return errors.New("用户不存在")
	}
	saltPassword := public.SaltPassword(admin.Salt, input.Password)
	if admin.Password != saltPassword {
		return errors.New("密码错误")
	}
	return nil
}
