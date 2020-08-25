package controller

import (
	"errors"
	"fmt"
	"github.com/e421083458/gateway/dao"
	"github.com/e421083458/gateway/dto"
	"github.com/e421083458/gateway/middleware"
	"github.com/e421083458/gateway/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

type ServiceController struct {
}

func ServiceRegister(group *gin.RouterGroup) {
	service := &ServiceController{}
	group.GET("https", service.Index)
	group.GET("http", service.Show)
	group.POST("https", service.AddHttp)
	group.PUT("https", service.Update)
	group.DELETE("https", service.Delete)
	group.GET("/stat", service.Stat)

	group.POST("/tcps", service.AddTcp)
	group.PUT("/tcps", service.UpdateTcp)
	group.POST("/grpcs", service.AddGrpc)
	group.PUT("/grpcs", service.UpdateGrpc)
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
// @Router /service/https [get]
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
			LoadType:    item.LoadType,
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

// ServiceAddHttp godoc
// @Summary 添加http服务
// @Description 服务列表
// @Tags 服务管理
// @ID /service/add_http
// @Accept json
// @Produce json
// @Param body body dto.ServiceAddHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/https [post]
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

	accessControl := &dao.AccessControl{
		ServiceID:         service.Id,
		OpenAuth:          input.OpenAuth,
		BlackList:         input.BlackList,
		WhiteList:         input.WhiteList,
		ClientIPFlowLimit: input.ClientipFlowLimit,
		ServiceFlowLimit:  input.ServiceFlowLimit,
	}
	if err := accessControl.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2007, err)
		return
	}

	loadbalance := &dao.LoadBalance{
		ServiceID:              service.Id,
		RoundType:              input.RoundType,
		IpList:                 input.IpList,
		WeightList:             input.WeightList,
		UpstreamConnectTimeout: input.UpstreamConnectTimeout,
		UpstreamHeaderTimeout:  input.UpstreamHeaderTimeout,
		UpstreamIdleTimeout:    input.UpstreamIdleTimeout,
		UpstreamMaxIdle:        input.UpstreamMaxIdle,
	}
	if err := loadbalance.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2008, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(ctx, "")
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
// @Router /service/https [delete]
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

// ServiceAddHttp godoc
// @Summary 修改http服务
// @Description 修改http服务
// @Tags 服务管理
// @ID /service/update
// @Accept json
// @Produce json
// @Param body body dto.ServiceUpdateHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/https [put]
func (c *ServiceController) Update(ctx *gin.Context) {
	input := &dto.ServiceAddHTTPInput{}
	if err := input.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	if len(strings.Split(input.IpList, ",")) != len(strings.Split(input.WeightList, ",")) {
		middleware.ResponseError(ctx, 2001, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}
	tx = tx.Begin()

	service := &dao.ServiceInfo{ServiceName: input.ServiceName}
	if err = service.Find(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2003, errors.New("服务不存在"))
		return
	}

	serviceDetail, err := service.ServiceDetail(ctx, tx)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2004, errors.New("服务不存在"))
		return
	}

	info := serviceDetail.Info
	info.ServiceDesc = input.ServiceDesc
	if err := info.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2005, err)
		return
	}

	httpRule := serviceDetail.HTTPRule
	httpRule.NeedHttps = input.NeedHttps
	httpRule.NeedStripUri = input.NeedStripUri
	httpRule.NeedWebsocket = input.NeedWebsocket
	httpRule.UrlRewrite = input.UrlRewrite
	httpRule.HeaderTransfor = input.HeaderTransfor
	if err := httpRule.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2006, err)
		return
	}

	accessControl := serviceDetail.AccessControl
	accessControl.OpenAuth = input.OpenAuth
	accessControl.BlackList = input.BlackList
	accessControl.WhiteList = input.WhiteList
	accessControl.ClientIPFlowLimit = input.ClientipFlowLimit
	accessControl.ServiceFlowLimit = input.ServiceFlowLimit
	if err := accessControl.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2007, err)
		return
	}

	loadbalance := serviceDetail.LoadBalance
	loadbalance.RoundType = input.RoundType
	loadbalance.IpList = input.IpList
	loadbalance.WeightList = input.WeightList
	loadbalance.UpstreamConnectTimeout = input.UpstreamConnectTimeout
	loadbalance.UpstreamHeaderTimeout = input.UpstreamHeaderTimeout
	loadbalance.UpstreamIdleTimeout = input.UpstreamIdleTimeout
	loadbalance.UpstreamMaxIdle = input.UpstreamMaxIdle
	if err := loadbalance.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2008, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(ctx, "")
}

// ServiceList godoc
// @Summary 服务详情
// @Description 服务详情
// @Tags 服务管理
// @ID /service/detail
// @Accept json
// @Produce json
// @Param id query string true "服务ID"
// @Success 200 {object} middleware.Response{data=dao.ServiceDetail} "success"
// @Router /service/http [get]
func (c *ServiceController) Show(ctx *gin.Context) {
	var (
		err    error
		tx     *gorm.DB
		detail *dao.ServiceDetail
	)

	input := &dto.ServiceDeleteInput{}
	if err = input.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	if tx, err = lib.GetGormPool("default"); err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	//读取基本信息
	service := &dao.ServiceInfo{Id: input.ID}
	if service.Find(ctx, tx) != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}
	if detail, err = service.ServiceDetail(ctx, tx); err != nil {
		middleware.ResponseError(ctx, 2003, err)
		return
	}
	middleware.ResponseSuccess(ctx, detail)
}

// ServiceStat godoc
// @Summary 服务统计
// @Description 服务统计
// @Tags 服务管理
// @ID /service/stat
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} middleware.Response{data=dto.ServiceStatOutput} "success"
// @Router /service/stat [get]
func (c *ServiceController) Stat(ctx *gin.Context) {
	var (
		err error
		//tx            *gorm.DB
		todayList     []int64
		yesterdayList []int64
	)

	input := &dto.ServiceDeleteInput{}
	if err = input.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	//读取基本信息
	/*if tx, err = lib.GetGormPool("default"); err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}
	info := &dao.ServiceInfo{Id: input.ID}
	if detail, err = info.ServiceDetail(ctx, tx); err != nil {
		middleware.ResponseError(ctx, 2003, err)
		return
	}*/
	for i := 0; i < time.Now().Hour(); i++ {
		todayList = append(todayList, 0)
	}
	for i := 0; i < 24; i++ {
		yesterdayList = append(yesterdayList, 0)
	}
	middleware.ResponseSuccess(ctx, &dto.ServiceStatOutput{
		Today:     todayList,
		Yesterday: yesterdayList,
	})
}

// ServiceAddHttp godoc
// @Summary tcp服务添加
// @Description tcp服务添加
// @Tags 服务管理
// @ID /service/add_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddTcpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/tcps [post]
func (c *ServiceController) AddTcp(ctx *gin.Context) {
	var (
		err error
	)

	input := &dto.ServiceAddTcpInput{}
	if err = input.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}
	// 验证 service_name 是否被占用
	info := &dao.ServiceInfo{
		ServiceName: input.ServiceName,
		IsDelete:    0,
	}
	if err = info.Find(ctx, lib.GORMDefaultPool); err == nil {
		middleware.ResponseError(ctx, 2002, errors.New("服务名被占用，请重新输入"))
		return
	}
	// 验证端口是否被占用?
	tcpRuleSearch := &dao.TcpRule{
		Port: input.Port,
	}
	if err := tcpRuleSearch.Find(ctx, lib.GORMDefaultPool); err == nil {
		middleware.ResponseError(ctx, 2003, errors.New("服务端口被占用，请重新输入"))
		return
	}
	grpcRuleSearch := &dao.GrpcRule{
		Port: input.Port,
	}
	if err := grpcRuleSearch.Find(ctx, lib.GORMDefaultPool); err == nil {
		middleware.ResponseError(ctx, 2004, errors.New("服务端口被占用，请重新输入"))
		return
	}

	//ip与权重数量一致
	if len(strings.Split(input.IpList, ",")) != len(strings.Split(input.WeightList, ",")) {
		middleware.ResponseError(ctx, 2005, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()
	info = &dao.ServiceInfo{
		LoadType:    public.LoadTypeTCP,
		ServiceName: input.ServiceName,
		ServiceDesc: input.ServiceDesc,
	}
	if err := info.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2006, err)
		return
	}
	loadBalance := &dao.LoadBalance{
		ServiceID:  info.Id,
		RoundType:  input.RoundType,
		IpList:     input.IpList,
		WeightList: input.WeightList,
		ForbidList: input.ForbidList,
	}
	if err := loadBalance.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2007, err)
		return
	}

	httpRule := &dao.TcpRule{
		ServiceID: info.Id,
		Port:      input.Port,
	}
	if err := httpRule.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2008, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         info.Id,
		OpenAuth:          input.OpenAuth,
		BlackList:         input.BlackList,
		WhiteList:         input.WhiteList,
		WhiteHostName:     input.WhiteHostName,
		ClientIPFlowLimit: input.ClientIPFlowLimit,
		ServiceFlowLimit:  input.ServiceFlowLimit,
	}
	if err := accessControl.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2009, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(ctx, "")
	return
}

// ServiceUpdateTcp godoc
// @Summary tcp服务更新
// @Description tcp服务更新
// @Tags 服务管理
// @ID /service/service_update_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateTcpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/tcps [put]
func (c *ServiceController) UpdateTcp(ctx *gin.Context) {
	params := &dto.ServiceUpdateTcpInput{}
	if err := params.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(ctx, 2002, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()

	service := &dao.ServiceInfo{
		Id: params.ID,
	}
	detail, err := service.ServiceDetail(ctx, lib.GORMDefaultPool)
	if err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}

	info := detail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2003, err)
		return
	}

	loadBalance := &dao.LoadBalance{}
	if detail.LoadBalance != nil {
		loadBalance = detail.LoadBalance
	}
	loadBalance.ServiceID = info.Id
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.ForbidList = params.ForbidList
	if err := loadBalance.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2004, err)
		return
	}

	tcpRule := &dao.TcpRule{}
	if detail.TCPRule != nil {
		tcpRule = detail.TCPRule
	}
	tcpRule.ServiceID = info.Id
	tcpRule.Port = params.Port
	if err := tcpRule.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2005, err)
		return
	}

	accessControl := &dao.AccessControl{}
	if detail.AccessControl != nil {
		accessControl = detail.AccessControl
	}
	accessControl.ServiceID = info.Id
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.WhiteHostName = params.WhiteHostName
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2006, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(ctx, "")
	return
}

// ServiceAddHttp godoc
// @Summary grpc服务添加
// @Description grpc服务添加
// @Tags 服务管理
// @ID /service/add_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddGrpcInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/grpcs [post]
func (c *ServiceController) AddGrpc(ctx *gin.Context) {
	params := &dto.ServiceAddGrpcInput{}
	if err := params.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	//验证 service_name 是否被占用
	infoSearch := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		IsDelete:    0,
	}
	if err := infoSearch.Find(ctx, lib.GORMDefaultPool); err == nil {
		middleware.ResponseError(ctx, 2002, errors.New("服务名被占用，请重新输入"))
		return
	}

	//验证端口是否被占用?
	tcpRuleSearch := &dao.TcpRule{
		Port: params.Port,
	}
	if err := tcpRuleSearch.Find(ctx, lib.GORMDefaultPool); err == nil {
		middleware.ResponseError(ctx, 2003, errors.New("服务端口被占用，请重新输入"))
		return
	}
	grpcRuleSearch := &dao.GrpcRule{
		Port: params.Port,
	}
	if err := grpcRuleSearch.Find(ctx, lib.GORMDefaultPool); err == nil {
		middleware.ResponseError(ctx, 2004, errors.New("服务端口被占用，请重新输入"))
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(ctx, 2005, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()
	info := &dao.ServiceInfo{
		LoadType:    public.LoadTypeGRPC,
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := info.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2006, err)
		return
	}

	loadBalance := &dao.LoadBalance{
		ServiceID:  info.Id,
		RoundType:  params.RoundType,
		IpList:     params.IpList,
		WeightList: params.WeightList,
		ForbidList: params.ForbidList,
	}
	if err := loadBalance.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2007, err)
		return
	}

	grpcRule := &dao.GrpcRule{
		ServiceID:      info.Id,
		Port:           params.Port,
		HeaderTransfor: params.HeaderTransfor,
	}
	if err := grpcRule.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2008, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         info.Id,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		WhiteHostName:     params.WhiteHostName,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2009, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(ctx, "")
	return
}

// ServiceUpdateTcp godoc
// @Summary grpc服务更新
// @Description grpc服务更新
// @Tags 服务管理
// @ID /service/update_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateGrpcInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/grpcs [put]
func (c *ServiceController) UpdateGrpc(ctx *gin.Context) {
	params := &dto.ServiceUpdateGrpcInput{}
	if err := params.GetValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(ctx, 2002, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()

	service := &dao.ServiceInfo{
		Id: params.ID,
	}
	detail, err := service.ServiceDetail(ctx, lib.GORMDefaultPool)
	if err != nil {
		middleware.ResponseError(ctx, 2003, err)
		return
	}

	info := detail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2004, err)
		return
	}

	loadBalance := &dao.LoadBalance{}
	if detail.LoadBalance != nil {
		loadBalance = detail.LoadBalance
	}
	loadBalance.ServiceID = info.Id
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.ForbidList = params.ForbidList
	if err := loadBalance.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2005, err)
		return
	}

	grpcRule := &dao.GrpcRule{}
	if detail.GRPCRule != nil {
		grpcRule = detail.GRPCRule
	}
	grpcRule.ServiceID = info.Id
	//grpcRule.Port = params.Port
	grpcRule.HeaderTransfor = params.HeaderTransfor
	if err := grpcRule.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2006, err)
		return
	}

	accessControl := &dao.AccessControl{}
	if detail.AccessControl != nil {
		accessControl = detail.AccessControl
	}
	accessControl.ServiceID = info.Id
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.WhiteHostName = params.WhiteHostName
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(ctx, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(ctx, 2007, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(ctx, "")
	return
}
