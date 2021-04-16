package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhyeah/gin-autoreg/vo"
)

// ControllerMap 用于controller注册
var ControllerMap map[string]interface{} = map[string]interface{}{}

// DefaultController 默认路由controller
type DefaultController struct {
}

// MethodNotFound 方法未找到
func (controller *DefaultController) MethodNotFound(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, vo.GeneralResponse{
		Code:    http.StatusNotFound,
		Message: "API is not exist",
	})
}
