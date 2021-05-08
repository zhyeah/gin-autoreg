package autoroute

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/zhyeah/gin-autoreg/controller"
	"github.com/zhyeah/gin-autoreg/exception"
	"github.com/zhyeah/gin-autoreg/param"
	"github.com/zhyeah/gin-autoreg/tag"
	"github.com/zhyeah/gin-autoreg/util"
	"github.com/zhyeah/gin-autoreg/vo"
)

const (
	Get    = "GET"
	Post   = "POST"
	Put    = "PUT"
	Delete = "DELETE"
)

const (
	TagFieldUrl    = "url"
	TagFieldMethod = "method"
	TagFieldFunc   = "func"
	TagFieldAuth   = "auth"
)

// AutoRouteConfig regitster route automatically
type AutoRouteConfig struct {
	Engine          *gin.Engine
	BaseUrl         string
	ResponseHandler func(ctx *gin.Context, exp *exception.HTTPException, data interface{})
	OAAuth          func(ctx *gin.Context)
}

var autoRouter *AutoRouter

type AutoRouter struct {
	AutoRouteConfig *AutoRouteConfig
	TagManager      *tag.Manager
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

			field := typ.Field(i)
			err := router.registerController(engine, v, &field)
			if err != nil {
				httpRequestTag := typ.Field(i).Tag.Get("httprequest")
				return fmt.Errorf(k + " register api " + httpRequestTag + " failed, err: " + err.Error())
			}
		}
	}

	return nil
}

func (router *AutoRouter) registerController(engine *gin.RouterGroup, ctrl interface{}, field *reflect.StructField) error {
	vals := strings.Split(field.Tag.Get("httprequest"), ";")
	if len(vals) < 3 {
		return errors.New("the arguments should contains at least 3 parameters: url, method, func")
	}

	tagMap := convertTagArrayToMap(vals)
	url, ok := tagMap[TagFieldUrl]
	if !ok {
		return errors.New("the tag field should contains 'url'")
	}
	method, ok := tagMap[TagFieldMethod]
	if !ok {
		return errors.New("the tag field should contains 'method'")
	}
	function, ok := tagMap[TagFieldFunc]
	if !ok {
		return errors.New("the tag field should contains 'func'")
	}
	needAuth, ok := tagMap[TagFieldAuth]
	if !ok {
		needAuth = "true"
	}

	args := []interface{}{url}

	// auth check
	if needAuth == "true" && router.AutoRouteConfig.OAAuth != nil {
		args = append(args, router.AutoRouteConfig.OAAuth)
	}

	// Pre-Handlers
	router.RegisterTagHandlers(field, &args, router.TagManager.GetPreHandlers())

	// http handler
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
			router.AutoRouteConfig.ResponseHandler(ctx, &exception.HTTPException{
				Code:    http.StatusBadRequest,
				Message: err.(error).Error(),
				Err:     err.(error),
			}, nil)
			return
		}

		rets := util.ReflectInvokeMethod(ctrl, function, args...)
		var data interface{} = nil
		if len(rets) == 0 {
			return
		} else if len(rets) == 1 {
			data = nil
			err = rets[0]
		} else {
			data = rets[0]
			err = rets[1]
		}

		if err == nil {
			router.AutoRouteConfig.ResponseHandler(ctx, nil, data)
		} else if reflect.TypeOf(err).Elem().Name() == "HTTPException" {
			httpException := err.(*exception.HTTPException)
			router.AutoRouteConfig.ResponseHandler(ctx, &exception.HTTPException{
				Code:    httpException.Code,
				Message: httpException.Message,
				Err:     httpException,
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

	// Post-Handlers
	router.RegisterTagHandlers(field, &args, router.TagManager.GetPostHandlers())

	util.ReflectInvokeMethod(engine, method, args...)
	return nil
}

// RegisterTagHandlers register tag handlers
func (router *AutoRouter) RegisterTagHandlers(field *reflect.StructField, args *[]interface{}, handlers map[string]tag.Handler) {
	if len(handlers) == 0 {
		return
	}

	handlersArray := make([]tag.Handler, 0)
	orderMap := make(map[int]string)
	for k, v := range handlers {
		tag := (*field).Tag.Get(k)
		if tag == "" {
			continue
		}

		handlersArray = append(handlersArray, v)
		orderMap[v.GetOrder()-1] = tag
	}

	// sort by 'GetOrder()'
	sort.SliceStable(handlersArray, func(i, j int) bool {
		return handlersArray[i].GetOrder() < handlersArray[j].GetOrder()
	})

	// add handlers to 'args'
	for i := range handlersArray {
		*args = append(*args, func(ctx *gin.Context) {
			tagValue, _ := orderMap[i]
			result := handlersArray[i].Handle(tagValue, ctx)
			if result.Code == tag.FailedAndStop {
				ctx.Abort()
			}
		})
	}
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
			autoRouter = &AutoRouter{
				TagManager: tag.GetManager(),
			}
		})
	}
	return autoRouter
}
