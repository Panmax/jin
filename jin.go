package jin

import "sync"

var (
	default404Body   = []byte("404 page not found")
	default405Body   = []byte("405 method not allowed")
	defaultAppEngine bool
)

type HandlerFunc func(*Context)

type HandlerChain []HandlerFunc

func (c HandlerChain) Last() HandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

type RouteInfo struct {
	Method      string
	Path        string
	Handler     string
	HandlerFunc HandlerFunc
}

type RoutesInfo []RouteInfo

type Engine struct {
	RouterGroup

	RedirectTrailingSlash bool

	RedirectFixedPath bool

	HandleMethodNotAllowed bool
	ForwardedByClientIP    bool

	AppEngine bool

	UseRawPath bool

	UnescapePathValues bool

	MaxMultipartMemory int64

	delims           render.Delims
	secureJsonPrefix string
	HTMLRender       render.HTMLRender
	FuncMap          template.FuncMap
	allNoRoute       HandlerChain
	allNoMethod      HandlerChain
	noRoute          HandlerChain
	noMethod         HandlerChain
	pool             sync.Pool
	trees            methodTrees
}

var _ IRouter = &Engine{} // 确保 Engine 实现 IRouter 接口
