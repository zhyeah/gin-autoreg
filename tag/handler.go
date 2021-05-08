package tag

import "github.com/gin-gonic/gin"

// 处理结果
const (
	Success       = 0
	FailedButGoOn = 1
	FailedAndStop = 2
)

// HandleResult 处理结果对象
type HandleResult struct {
	Code int
}

// Handler 自定义标签处理器接口
type Handler interface {
	GetOrder() int
	Handle(tagValue string, ctx *gin.Context) *HandleResult
}
