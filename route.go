package autoroute

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime/debug"
	"sort"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/zhyeah/gin-autoreg/controller"
	"github.com/zhyeah/gin-autoreg/data"
	"github.com/zhyeah/gin-autoreg/exception"
	"github.com/zhyeah/gin-autoreg/intercepter"
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
	TagFieldAuthor = "author"
)

// AutoRouteConfig regitster route automatically
type AutoRouteConfig struct {
	Engine          *gin.Engine
	BaseUrl         string
	ResponseHandler func(ctx *gin.Context, exp *exception.HTTPException, data interface{})
	OAAuth          func(ctx *gin.Context, forceCheck bool)
}

var autoRouter *AutoRouter

type AutoRouter struct {
	AutoRouteConfig   *AutoRouteConfig
	TagManager        *tag.Manager
	OnStartActions    []func(*data.RouterContext)
	Context           *data.RouterContext
	OnFinishedActions []func(*data.RouterContext)
}

// AddStartAction 添加启动action
func (router *AutoRouter) AddStartAction(startAction func(*data.RouterContext)) {
	if router.OnStartActions == nil {
		router.OnStartActions = make([]func(*data.RouterContext), 0)
	}
	router.OnStartActions = append(router.OnStartActions, startAction)
}

// AddStartAction 添加启动action
func (router *AutoRouter) AddFinishedAction(finishedAction func(*data.RouterContext)) {
	if router.OnFinishedActions == nil {
		router.OnFinishedActions = make([]func(*data.RouterContext), 0)
	}
	router.OnFinishedActions = append(router.OnFinishedActions, finishedAction)
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
	if router.Context == nil {
		router.Context = &data.RouterContext{
			HTTPMap: make(map[string]*data.HTTPRequest),
		}
	}

	// default route
	var defaultController controller.DefaultController
	config.Engine.NoRoute(defaultController.MethodNotFound)

	// base url
	route := config.Engine.Group(config.BaseUrl)

	// on start
	router.onStart()

	// register route of each controller
	err := router.registerEachController(route)
	if err != nil {
		return err
	}

	// on end
	router.onFinished()

	return nil
}

// onStart boot action
func (router *AutoRouter) onStart() {
	if router.OnStartActions != nil {
		for i := range router.OnStartActions {
			router.OnStartActions[i](router.Context)
		}
	}
}

// onStart boot action
func (router *AutoRouter) onFinished() {
	if router.OnStartActions != nil {
		for i := range router.OnFinishedActions {
			router.OnFinishedActions[i](router.Context)
		}
	}
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

	httpRequest, err := convertTag(ctrl, vals)
	if err != nil {
		return err
	}
	// fill http request into context map
	router.Context.HTTPMap[httpRequest.URL] = httpRequest

	args := []interface{}{httpRequest.URL}

	// auth check
	if router.AutoRouteConfig.OAAuth != nil {
		args = append(args, func(ctx *gin.Context) {
			router.AutoRouteConfig.OAAuth(ctx, httpRequest.Auth)
		})
	}

	// Pre-Intercepters
	intercepterManager := intercepter.GetIntercepterManager()
	preInters := intercepterManager.GetPreIntercepters()
	for i := range preInters {
		args = append(args, preInters[i])
	}

	// Pre-Handlers
	router.RegisterTagHandlers(field, &args, router.TagManager.GetPreHandlers())

	// http handler
	args = append(args, func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(string(debug.Stack())) // just for debug
				router.AutoRouteConfig.ResponseHandler(ctx, &exception.HTTPException{
					Code:    http.StatusInternalServerError,
					Message: fmt.Sprintf("error: %v", err),
					Err:     nil,
				}, nil)
				ctx.Abort()
			}
		}()

		var err interface{} = nil
		args, err := param.ResolveParams(ctrl, httpRequest.Func, ctx)
		ctx.Set("args", args)
		if err != nil {
			router.AutoRouteConfig.ResponseHandler(ctx, &exception.HTTPException{
				Code:    http.StatusBadRequest,
				Message: err.(error).Error(),
				Err:     err.(error),
			}, nil)
			return
		}

		rets := util.ReflectInvokeMethod(ctrl, httpRequest.Func, args...)
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

	// Post-inters
	postInters := intercepterManager.GetPreIntercepters()
	for i := range postInters {
		args = append(args, postInters[i])
	}

	util.ReflectInvokeMethod(engine, httpRequest.Method, args...)
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

func convertTag(ctrl interface{}, tags []string) (*data.HTTPRequest, error) {

	tagMap := make(map[string]string)

	for _, tag := range tags {
		params := strings.Split(tag, "=")
		if len(params) < 2 {
			continue
		}
		tagMap[params[0]] = params[1]
	}

	url, ok := tagMap[TagFieldUrl]
	if !ok {
		return nil, errors.New("the tag field should contains 'url'")
	}
	method, ok := tagMap[TagFieldMethod]
	if !ok {
		return nil, errors.New("the tag field should contains 'method'")
	}
	function, ok := tagMap[TagFieldFunc]
	if !ok {
		return nil, errors.New("the tag field should contains 'func'")
	}
	needAuth, ok := tagMap[TagFieldAuth]
	if !ok {
		needAuth = "true"
	}
	author, ok := tagMap[TagFieldAuthor]
	if !ok {
		author = ""
	}
	dataStr, err := param.ResolvePostDataJson(ctrl, method)
	if err != nil {
		return nil, err
	}

	return &data.HTTPRequest{
		URL:    url,
		Method: method,
		Func:   function,
		Auth:   util.ConvertStringToBoolDefault(needAuth, true),
		Author: author,
		Data:   dataStr,
	}, nil
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
