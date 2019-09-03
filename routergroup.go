package jin

import "net/http"

type IRouter interface {
	IRoutes
	Group(string, ...HandlerFunc) *RouterGroup
}

// IRoutes 定义所有路由处理接口
type IRoutes interface {
	Use(...HandlerFunc) IRoutes

	Handle(string, string, ...HandlerFunc) IRoutes
	Any(string, ...HandlerFunc) IRoutes
	GET(string, ...HandlerFunc) IRoutes
	POST(string, ...HandlerFunc) IRoutes
	DELETE(string, ...HandlerFunc) IRoutes
	PUT(string, ...HandlerFunc) IRoutes
	PATCH(string, ...HandlerFunc) IRoutes
	OPTIONS(string, ...HandlerFunc) IRoutes
	HEAD(string, ...HandlerFunc) IRoutes

	StaticFile(string, string) IRoutes
	Static(string, string) IRoutes
	StaticFS(string, system http.FileSystem) IRoutes
}

// RouterGroup 用来内部配置路由，一个 RouterGroup 关联一个
// 前缀和一组处理器（中间件）
type RouterGroup struct {
	Handlers HandlerChain
	basePath string
	engine   *Engine
	root     bool
}

var _ IRouter = &RouterGroup{} // 确保 RouterGroup 实现 IRouter 接口

func (group *RouterGroup) Use(middleware ...HandlerFunc) IRoutes {
	group.Handlers = append(group.Handlers, middleware...)
	return group.returnObj()
}

func (group *RouterGroup) returnObj() IRoutes {
	if group.root {
		return group.engine
	}
	return group
}
