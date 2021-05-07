package tag

import "sync"

var managerOnce sync.Once
var manager *Manager

// Manager 标签管理者
type Manager struct {
	preHandlers  map[string]Handler
	postHandlers map[string]Handler
}

// AddPreHandler 添加前置处理器
func (m *Manager) AddPreHandler(tagName string, handler Handler) {
	m.preHandlers[tagName] = handler
}

// GetPreHandlers 获取前置处理器
func (m *Manager) GetPreHandlers() map[string]Handler {
	return m.preHandlers
}

// AddPostHandler 添加后置处理器
func (m *Manager) AddPostHandler(tagName string, handler Handler) {
	m.postHandlers[tagName] = handler
}

// GetPostHandlers 获取后置处理器
func (m *Manager) GetPostHandlers() map[string]Handler {
	return m.postHandlers
}

// GetManager 获取标签管理器
func GetManager() *Manager {
	if manager == nil {
		managerOnce.Do(func() {
			manager = &Manager{
				preHandlers:  make(map[string]Handler),
				postHandlers: make(map[string]Handler),
			}
		})
	}
	return manager
}
