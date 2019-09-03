package jin

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
}

var _ IRouter = &Engine{} // 确保 Engine 实现 IRouter 接口
