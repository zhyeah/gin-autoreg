package intercepter

import (
	"sync"

	"github.com/gin-gonic/gin"
)

var intercepterManagerOnce sync.Once
var intercepterManager *IntercepterManager

// IntercepterManager 拦截器管理者
type IntercepterManager struct {
	preIntercepters  []func(*gin.Context)
	postIntercepters []func(*gin.Context)
}

// AddPreIntercepters 添加拦截器
func (manager *IntercepterManager) AddPreIntercepters(handler func(*gin.Context)) {
	manager.preIntercepters = append(manager.preIntercepters, handler)
}

// GetPreIntercepters 获取拦截器
func (manager *IntercepterManager) GetPreIntercepters() []func(*gin.Context) {
	return manager.preIntercepters
}

// AddPostIntercepters 添加拦截器
func (manager *IntercepterManager) AddPostIntercepters(handler func(*gin.Context)) {
	manager.postIntercepters = append(manager.postIntercepters, handler)
}

// GetPostIntercepters 获取拦截器
func (manager *IntercepterManager) GetPostIntercepters() []func(*gin.Context) {
	return manager.postIntercepters
}

// GetIntercepterManager 获取拦截器管理器
func GetIntercepterManager() *IntercepterManager {
	if intercepterManager == nil {
		intercepterManagerOnce.Do(func() {
			intercepterManager = &IntercepterManager{
				preIntercepters:  make([]func(*gin.Context), 0),
				postIntercepters: make([]func(*gin.Context), 0),
			}
		})
	}
	return intercepterManager
}
