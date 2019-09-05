package jin

import (
	"math"
	"net/http"
	"net/url"
)

const abortIndex int8 = math.MaxInt8 / 2

// Context 是 jin 里最重要的一部分。它允许我们在中间件之间传递变量。
// 管理流，验证请求参数，渲染响应的 Json
type Context struct {
	writermem responseWriter
	Request   *http.Request
	Writer    ResponseWriter

	Params   Params
	handlers HandlerChain
	index    int8
	fullPath string

	engine *Engine

	Keys map[string]interface{}

	Errors errorMsgs

	Accepted []string

	queryCache url.Values

	formCache url.Values
}

func (c *Context) Param(key string) string {
	return c.Params.ByName(key)
}

func (c *Context) File(filepath string) {
	http.ServeFile(c.Writer, c.Request, filepath)
}
