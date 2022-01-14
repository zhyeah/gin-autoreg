package intercepter

import (
	"sync"

	"github.com/gin-gonic/gin"
)

// PanicInfo panic信息
type PanicInfo struct {
	RecoverContent interface{}
	ExceptionStack string
}

var globalPanicIntercepter *GlobalPanicIntercepter
var globalPanicIntercepterOnce sync.Once

// GlobalPanicIntercepter 全局panic拦截器
type GlobalPanicIntercepter struct {
	panicHandleFuncs []func(ctx *gin.Context, panicInfo *PanicInfo)
}

// AppendPanicIntercepter 追加panic拦截器
func (inter *GlobalPanicIntercepter) AppendPanicIntercepter(handleFunc func(ctx *gin.Context, panicInfo *PanicInfo)) {
	if inter.panicHandleFuncs == nil {
		inter.panicHandleFuncs = make([]func(ctx *gin.Context, panicInfo *PanicInfo), 0)
	}
	inter.panicHandleFuncs = append(inter.panicHandleFuncs, handleFunc)
}

// GetPanicHandleFuncs 获取全局panic处理器函数
func (inter *GlobalPanicIntercepter) GetPanicHandleFuncs() []func(ctx *gin.Context, panicInfo *PanicInfo) {
	return inter.panicHandleFuncs
}

// GetGlobalPanicIntercepter 获取全局panic拦截器
func GetGlobalPanicIntercepter() *GlobalPanicIntercepter {
	if globalPanicIntercepter == nil {
		globalPanicIntercepterOnce.Do(func() {
			globalPanicIntercepter = &GlobalPanicIntercepter{}
		})
	}
	return globalPanicIntercepter
}
