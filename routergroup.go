package jin

import (
	"net/http"
	"path"
	"regexp"
	"strings"
)

// IRouter 定义了所有路由处理接口包括单个和群组路由
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
	StaticFS(string, http.FileSystem) IRoutes
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

// Use 添加中间件到组
func (group *RouterGroup) Use(middleware ...HandlerFunc) IRoutes {
	group.Handlers = append(group.Handlers, middleware...)
	return group.returnObj()
}

// Group 创建一个新的路由组，你应该将所有用共同中间件或相同前缀的路由加入进来
// 比如，用了同样中间件来实现授权的路由可以放在同一组
func (group *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		Handlers: group.combineHandlers(handlers),
		basePath: group.calculateAbsolutePath(relativePath),
		engine:   group.engine,
	}
}

func (group *RouterGroup) BasePath() string {
	return group.basePath
}

func (group *RouterGroup) handle(httpMethod, relativePath string, handlers HandlerChain) IRoutes {
	absolutePath := group.calculateAbsolutePath(relativePath)
	handlers = group.combineHandlers(handlers)
	group.engine.addRoute(httpMethod, absolutePath, handlers)
	return group.returnObj()
}

func (group *RouterGroup) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) IRoutes {
	if matches, err := regexp.MatchString("^[A-Z]+$", httpMethod); !matches || err != nil {
		panic("http method " + httpMethod + " is not valid")
	}
	return group.handle(httpMethod, relativePath, handlers)
}

func (group *RouterGroup) POST(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("POST", relativePath, handlers)
}

func (group *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("GET", relativePath, handlers)
}

func (group *RouterGroup) DELETE(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("DELETE", relativePath, handlers)
}

func (group *RouterGroup) PATCH(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("PATCH", relativePath, handlers)
}

func (group *RouterGroup) PUT(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("PUT", relativePath, handlers)
}

func (group *RouterGroup) OPTIONS(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("OPTIONS", relativePath, handlers)
}

func (group *RouterGroup) HEAD(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("HEAD", relativePath, handlers)
}

func (group *RouterGroup) Any(relativePath string, handlers ...HandlerFunc) IRoutes {
	group.handle("GET", relativePath, handlers)
	group.handle("POST", relativePath, handlers)
	group.handle("PUT", relativePath, handlers)
	group.handle("PATCH", relativePath, handlers)
	group.handle("HEAD", relativePath, handlers)
	group.handle("OPTIONS", relativePath, handlers)
	group.handle("DELETE", relativePath, handlers)
	group.handle("CONNECT", relativePath, handlers)
	group.handle("TRACH", relativePath, handlers)
	return group.returnObj()
}

func (group *RouterGroup) StaticFile(relativePath, filepath string) IRoutes {
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters can not be used when serving a static file")
	}
	handler := func(c *Context) {
		c.File(filepath)
	}
	group.GET(relativePath, handler)
	group.HEAD(relativePath, handler)
	return group.returnObj()
}

func (group *RouterGroup) Static(relativePath, root string) IRoutes {
	return group.StaticFS(relativePath, Dir(root, false))
}

func (group *RouterGroup) StaticFS(relativePath string, fs http.FileSystem) IRoutes {
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters can not be used when serving a static folder")
	}
	handler := group.createStaticHandler(relativePath, fs)
	urlPattern := path.Join(relativePath, "/*filepath")

	group.GET(urlPattern, handler)
	group.HEAD(urlPattern, handler)
	return group.returnObj()
}

func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := group.calculateAbsolutePath(relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))

	return func(c *Context) {
		// fs.(*onlyFilesFs) 断言 fs 是 onlyFilesFs 类型
		if _, nolisting := fs.(*onlyFilesFs); nolisting {
			c.Writer.WriteHeader(http.StatusNotFound)
		}
		file := c.Param("filepath")
		if _, err := fs.Open(file); err != nil {
			c.Writer.WriteHeader(http.StatusNotFound)
			c.handlers = group.engine.noRoute

			c.index = -1
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Request)
	}

}

func (group *RouterGroup) combineHandlers(handlers HandlerChain) HandlerChain {
	finalSize := len(group.Handlers) + len(handlers)
	if finalSize >= int(abortIndex) {
		panic("too many handlers")
	}
	mergedHandlers := make(HandlerChain, finalSize)
	copy(mergedHandlers, group.Handlers)
	copy(mergedHandlers[len(group.Handlers):], handlers)
	return mergedHandlers
}

func (group *RouterGroup) calculateAbsolutePath(relativePath string) string {
	return joinPaths(group.basePath, relativePath)
}

func (group *RouterGroup) returnObj() IRoutes {
	if group.root {
		return group.engine
	}
	return group
}
