package dao

import (
	"github.com/e421083458/gateway/dto"
	"github.com/e421083458/gateway/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"sync"
	"time"
)

var (
	AppManagerHandler *AppManager
)

type AppManager struct {
	AppMap   map[string]*App
	AppSlice []*App
	Locker   sync.RWMutex
	init     sync.Once
	err      error
}

func NewAppManager() *AppManager {
	return &AppManager{
		AppMap:   map[string]*App{},
		AppSlice: []*App{},
		Locker:   sync.RWMutex{},
		init:     sync.Once{},
		err:      nil,
	}
}

func (a *AppManager) GetAppList() []*App {
	return a.AppSlice
}

func (a *AppManager) LoadOnce() (err error) {
	a.init.Do(func() {
		var (
			tx   *gorm.DB
			list []App
		)
		app := &App{}
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		if tx, err = lib.GetGormPool("default"); err != nil {
			a.err = err
			return
		}
		params := &dto.APPListInput{PageNo: 1, PageSize: 9999}
		if list, _, err = app.Select(ctx, tx, params); err != nil {
			a.err = err
			return
		}
		a.Locker.Lock()
		defer a.Locker.Unlock()

		for _, item := range list {
			a.AppMap[item.AppID] = &item
			a.AppSlice = append(a.AppSlice, &item)
		}
	})
	return a.err
}

type App struct {
	ID        int64     `json:"id" gorm:"primary_key"`
	AppID     string    `json:"app_id" gorm:"column:app_id" description:"租户id	"`
	Name      string    `json:"name" gorm:"column:name" description:"租户名称	"`
	Secret    string    `json:"secret" gorm:"column:secret" description:"密钥"`
	WhiteIPS  string    `json:"white_ips" gorm:"column:white_ips" description:"ip白名单，支持前缀匹配"`
	Qpd       int64     `json:"qpd" gorm:"column:qpd" description:"日请求量限制"`
	Qps       int64     `json:"qps" gorm:"column:qps" description:"每秒请求量限制"`
	CreatedAt time.Time `json:"create_at" gorm:"column:create_at" description:"添加时间	"`
	UpdatedAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	IsDelete  int8      `json:"is_delete" gorm:"column:is_delete" description:"是否已删除；0：否；1：是"`
}

func (a *App) TableName() string {
	return "gateway_app"
}

func (a *App) Find(ctx *gin.Context, db *gorm.DB) (err error) {
	return db.SetCtx(public.GetGinTraceContext(ctx)).Where(a).First(a).Error
}

func (a *App) Save(ctx *gin.Context, db *gorm.DB) (err error) {
	return db.SetCtx(public.GetGinTraceContext(ctx)).Save(a).Error
}

func (a *App) Select(ctx *gin.Context, db *gorm.DB, params *dto.APPListInput) (apps []App, total int64, err error) {
	offset := (params.PageNo - 1) * params.PageSize
	query := db.SetCtx(public.GetGinTraceContext(ctx))
	query = query.Table(a.TableName()).Select("*")
	query = query.Where("is_delete=?", 0)
	if params.Info != "" {
		query = query.Where(" (name like ? or app_id like ?)", "%"+params.Info+"%", "%"+params.Info+"%")
	}
	err = query.Limit(params.PageSize).Offset(offset).Order("id desc").Find(&apps).Error
	query.Limit(params.PageSize).Offset(offset).Count(&total)
	return
}
