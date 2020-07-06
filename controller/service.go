package controller

import (
	"errors"
	"fmt"
	"github.com/e421083458/gateway/dao"
	"github.com/e421083458/gateway/dto"
	"github.com/e421083458/gateway/middleware"
	"github.com/e421083458/gateway/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"strings"
)

type ServiceController struct {
}

func ServiceRegister(group *gin.RouterGroup) {
	service := &ServiceController{}
	group.GET("", service.Index)
	group.POST("", service.AddHttp)
	group.DELETE("", service.Delete)
}

// ServiceAddHttp godoc
// @Summary 添加http服务
// @Description 服务列表
// @Tags 服务管理
// @ID /service/add_http
// @Accept json
// @Produce json
// @Param body body dto.ServiceAddHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /services [post]
func (c *ServiceController) AddHttp(ctx *gin.Context) {
	input := &dto.ServiceAddHTTPInput{}
	if err := input.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	// 事务
	tx = tx.Begin()
	service := &dao.ServiceInfo{ServiceName: input.ServiceName}
	if err = service.Find(ctx, tx); err == nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2002, errors.New("服务已存在"))
		return
	}
	httpUrl := &dao.HttpRule{
		RuleType: input.RuleType,
		Rule:     input.Rule,
	}
	if err = httpUrl.Find(ctx, tx); err == nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2003, errors.New("服务接入前缀或域名已存在"))
		return
	}
	if len(strings.Split(input.IpList, "\n")) != len(strings.Split(input.WeightList, "\n")) {
		tx.Rollback()
		middleware.ResponseError(ctx, 2004, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	service = &dao.ServiceInfo{
		ServiceName: input.ServiceName,
		ServiceDesc: input.ServiceDesc,
	}
	if err := service.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2005, err)
		return
	}

	httpRule := &dao.HttpRule{
		ServiceID:      service.Id,
		RuleType:       input.RuleType,
		Rule:           input.Rule,
		NeedHttps:      input.NeedHttps,
		NeedStripUri:   input.NeedStripUri,
		NeedWebsocket:  input.NeedWebsocket,
		UrlRewrite:     input.UrlRewrite,
		HeaderTransfor: input.HeaderTransfor,
	}
	if err := httpRule.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2006, err)
		return
	}

	middleware.ResponseSuccess(ctx, "")
}

// ServiceList godoc
// @Summary 服务列表
// @Description 服务列表
// @Tags 服务管理
// @ID /service/index
// @Accept json
// @Produce json
// @Param info query string false "关键词"
// @Param page_no query int true "页码"
// @Param page_size query int true "每页条数"
// @Success 200 {object} middleware.Response{data=dto.ServiceListOutput} "success"
// @Router /services [get]
func (c *ServiceController) Index(ctx *gin.Context) {
	input := &dto.ServiceListInput{}
	if err := input.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}
	//基本信息
	serviceInfo := &dao.ServiceInfo{}
	list, total, err := serviceInfo.Page(ctx, tx, input)
	if err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}
	// 格式化输出
	outList := []dto.ServiceItemOutput{}
	for _, item := range list {
		serviceDetail, err := item.ServiceDetail(ctx, tx)
		if err != nil {
			middleware.ResponseError(ctx, 2003, err)
			return
		}

		serviceAddr := "unknown"
		clusterIp := lib.GetStringConf("base.cluster.cluster_ip")
		clusterPort := lib.GetStringConf("base.cluster.cluster_port")
		clusterSslIp := lib.GetStringConf("base.cluster.cluster_ssl_port")

		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL &&
			serviceDetail.HTTPRule.NeedHttps == 1 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIp, clusterSslIp, serviceDetail.HTTPRule.Rule)
		}
		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL &&
			serviceDetail.HTTPRule.NeedHttps == 0 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIp, clusterPort, serviceDetail.HTTPRule.Rule)
		}
		if serviceDetail.Info.LoadType == public.LoadTypeHTTP && serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypeDomain {
			serviceAddr = serviceDetail.HTTPRule.Rule
		}
		if serviceDetail.Info.LoadType == public.LoadTypeTCP {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIp, serviceDetail.TCPRule.Port)
		}
		if serviceDetail.Info.LoadType == public.LoadTypeGRPC {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIp, serviceDetail.GRPCRule.Port)
		}

		ips := serviceDetail.LoadBalance.GetIpListByMode()

		outItem := dto.ServiceItemOutput{
			Id:          item.Id,
			ServiceName: item.ServiceName,
			ServiceDesc: item.ServiceDesc,
			ServiceAddr: serviceAddr,
			Qps:         0,
			Qpd:         0,
			TotalNode:   len(ips),
		}
		outList = append(outList, outItem)
	}

	middleware.ResponseSuccess(ctx, &dto.ServiceListOutput{
		Total: total,
		List:  outList,
	})
}

// ServiceDelete godoc
// @Summary 删除服务
// @Description 删除服务
// @Tags 服务管理
// @ID /service/delete
// @Accept json
// @Produce json
// @Param id query string true "服务ID"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /services [delete]
func (c *ServiceController) Delete(ctx *gin.Context) {
	input := &dto.ServiceDeleteInput{}
	if err := input.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	info := &dao.ServiceInfo{
		Id: input.ID,
	}
	if err := info.Find(ctx, tx); err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}
	info.IsDelete = 1
	if err = info.Save(ctx, tx); err != nil {
		middleware.ResponseError(ctx, 2003, err)
		return
	}
	middleware.ResponseSuccess(ctx, "")
}
