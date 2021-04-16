package autoroute

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/zhyeah/gin-autoreg/controller"
	"github.com/zhyeah/gin-autoreg/exception"
	"github.com/zhyeah/gin-autoreg/param"
	"github.com/zhyeah/gin-autoreg/util"
	"github.com/zhyeah/gin-autoreg/vo"
)

const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
)

const (
	TAG_FIELD_URL    = "url"
	TAG_FIELD_METHOD = "method"
	TAG_FIELD_FUNC   = "func"
	TAG_FIELD_AUTH   = "auth"
)

// AutoRouteConfig 自动注入路由配置
type AutoRouteConfig struct {
	Engine          *gin.Engine
	BaseUrl         string
	ResponseHandler func(ctx *gin.Context, exp *exception.HTTPException, data interface{})
	OAAuth          func(ctx *gin.Context)
}

var autoRouter *AutoRouter

type AutoRouter struct {
	AutoRouteConfig *AutoRouteConfig
}

// RegisterRoute 注册路由
func (router *AutoRouter) RegisterRoute(config *AutoRouteConfig) error {
	router.AutoRouteConfig = config
	if config.ResponseHandler == nil {
		config.ResponseHandler = func(ctx *gin.Context, exp *exception.HTTPException, data interface{}) {
			if exp != nil {
				ctx.JSON(http.StatusOK, vo.GeneralResponse{
					Code:    exp.Code,
					Message: exp.Message,
				})
			} else {
				ctx.JSON(http.StatusOK, vo.GeneralResponse{
					Code:    0,
					Message: "",
					Data:    data,
				})
			}
		}
	}

	// 默认路由
	var defaultController controller.DefaultController
	config.Engine.NoRoute(defaultController.MethodNotFound)

	// 基础url
	route := config.Engine.Group(config.BaseUrl)

	// 注册各个controller的路由
	err := router.registerEachController(route)
	if err != nil {
		return err
	}

	return nil
}

func (router *AutoRouter) registerEachController(engine *gin.RouterGroup) error {

	for k, v := range controller.ControllerMap {
		typ := reflect.TypeOf(v).Elem()
		for i := 0; i < typ.NumField(); i++ {
			// route字段必须以route打头
			if strings.Index(typ.Field(i).Name, "route") != 0 {
				continue
			}
			httpRequestTag := typ.Field(i).Tag.Get("httprequest")
			err := router.registerController(engine, v, httpRequestTag)
			if err != nil {
				return fmt.Errorf(k + " register api " + httpRequestTag + " failed, err: " + err.Error())
			}
		}
	}

	return nil
}

func (router *AutoRouter) registerController(engine *gin.RouterGroup, ctrl interface{}, tag string) error {
	vals := strings.Split(tag, ";")
	if len(vals) < 3 {
		return errors.New("the arguments should contains at least 3 parameters: url, method, func")
	}

	tagMap := convertTagArrayToMap(vals)
	url, ok := tagMap[TAG_FIELD_URL]
	if !ok {
		return errors.New("the tag field should contains 'url'")
	}
	method, ok := tagMap[TAG_FIELD_METHOD]
	if !ok {
		return errors.New("the tag field should contains 'method'")
	}
	function, ok := tagMap[TAG_FIELD_FUNC]
	if !ok {
		return errors.New("the tag field should contains 'func'")
	}
	needAuth, ok := tagMap[TAG_FIELD_AUTH]
	if !ok {
		needAuth = "true"
	}

	args := []interface{}{url}

	// 对于需要检查权限的接口
	if needAuth == "true" && router.AutoRouteConfig.OAAuth != nil {
		args = append(args, router.AutoRouteConfig.OAAuth)
	}

	args = append(args, func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				router.AutoRouteConfig.ResponseHandler(ctx, &exception.HTTPException{
					Code:    http.StatusInternalServerError,
					Message: fmt.Sprintf("error: %v", err),
					Err:     nil,
				}, nil)
				ctx.Abort()
			}
		}()

		var err interface{} = nil
		args, err := param.ResolveParams(ctrl, function, ctx)
		if err != nil {
			// bizController.SendResponse(ctx, errno.New(errno.InvalidParams, err.(error)), nil)
			router.AutoRouteConfig.ResponseHandler(ctx, &exception.HTTPException{
				Code:    http.StatusBadRequest,
				Message: err.(error).Error(),
				Err:     err.(error),
			}, nil)
			return
		}

		rets := util.ReflectInvokeMethod(ctrl, function, args...)
		var data interface{} = nil
		if len(rets) == 1 {
			data = nil
			err = rets[0]
		} else {
			data = rets[0]
			err = rets[1]
		}

		if err == nil {
			router.AutoRouteConfig.ResponseHandler(ctx, nil, data)
		} else if reflect.TypeOf(err).Name() == "HTTPException" {
			httpException := err.(exception.HTTPException)
			router.AutoRouteConfig.ResponseHandler(ctx, &exception.HTTPException{
				Code:    httpException.Code,
				Message: httpException.Message,
				Err:     &httpException,
			}, nil)
		} else {
			err := err.(error)
			router.AutoRouteConfig.ResponseHandler(ctx, &exception.HTTPException{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
				Err:     err,
			}, nil)
		}
	})

	util.ReflectInvokeMethod(engine, method, args...)
	return nil
}

func convertTagArrayToMap(tags []string) map[string]string {

	retMap := make(map[string]string)

	for _, tag := range tags {
		params := strings.Split(tag, "=")
		if len(params) < 2 {
			continue
		}
		retMap[params[0]] = params[1]
	}

	return retMap
}

// GetAutoRouter 获取自动路由注册
func GetAutoRouter() *AutoRouter {
	if autoRouter == nil {
		var one sync.Once
		one.Do(func() {
			autoRouter = &AutoRouter{}
		})
	}
	return autoRouter
}
