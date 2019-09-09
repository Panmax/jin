package jin

import (
	"math"
	"net/http"
	"net/url"
	"strings"
)

const abortIndex int8 = math.MaxInt8 / 2

// Context 是 jin 里最重要的部分。它允许我们在中间件之间传递变量。
// 管理流，验证请求 JSON，渲染响应的 Json
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

func (c *Context) reset() {
	c.Writer = &c.writermem
	c.Params = c.Params[0:0]
	c.handlers = nil
	c.index = -1
	c.fullPath = ""
	c.Keys = nil
	c.Errors = c.Errors[0:0]
	c.Accepted = nil
	c.queryCache = nil
	c.formCache = nil
}

func (c *Context) Copy() *Context {
	var cp = *c
	cp.writermem.ResponseWriter = nil
	cp.Writer = &cp.writermem
	cp.index = abortIndex
	cp.Keys = map[string]interface{}{}
	for k, v := range c.Keys {
		cp.Keys[k] = v
	}
	paramCopy := make([]Param, len(cp.Params))
	copy(paramCopy, cp.Params)
	cp.Params = paramCopy
	return &cp
}

func (c *Context) HandlerName() string {
	return nameOfFunction(c.handlers.Last())
}

func (c *Context) HandlerNames() []string {
	hn := make([]string, len(c.handlers))
	for _, val := range c.handlers {
		hn = append(hn, nameOfFunction(val))
	}
	return hn
}

func (c *Context) Handler() HandlerFunc {
	return c.handlers.Last()
}

func (c *Context) FullPath() string {
	return c.fullPath
}

func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

func (c *Context) IsAborted() bool {
	return c.index >= abortIndex
}

func (c *Context) Abort() {
	c.index = abortIndex
}

func (c *Context) AbortWithStatus(code int) {
	c.Status(code)
	c.Writer.WriteHeaderNow()
	c.Abort()
}

func (c *Context) AbortWithStatusJSON(code int, jsonObj interface{}) {
	c.Abort()
	c.JSON(code, jsonObj)
}

func (c *Context) AbortWithError(code int, err error) *Error {
	c.AbortWithStatus(code)
	return c.Error(err)
}

func (c *Context) Error(err error) *Error {
	if err == nil {
		panic("err is nil")
	}

	parsedError, ok := err.(*Error)
	if !ok {
		parsedError = &Error{
			Err:  err,
			Type: ErrorTypePrivate,
		}
	}

	c.Errors = append(c.Errors, parsedError)
	return parsedError
}

// Set 专门用来为此上下文存储新的 键/值 对
// 如果之前未调用过，同样是懒初始化 c.Keys
func (c *Context) Set(key string, value interface{}) {
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[key] = value
}

func (c *Context) Get(key string) (value interface{}, exists bool) {
	value, exists = c.Keys[key]
	return
}

func (c *Context) MustGet(key string) interface{} {
	if val, exists := c.Get(key); exists {
		return val
	}
	panic("Key \"" + key + "\" does not exist")
}

func (c *Context) GetString(key string) (s string) {
	if val, ok := c.Get(key); ok && val != nil {
		s, _ = val.(string)
	}
	return
}

// TODO more get

func (c *Context) Param(key string) string {
	return c.Params.ByName(key)
}

func (c *Context) Query(key string) string {
	value, _ := c.GetQuery(key)
	return value
}

func (c *Context) DefaultQuery(key, defaultValue string) string {
	if value, ok := c.GetQuery(key); ok {
		return value
	}
	return defaultValue
}

func (c *Context) QueryArray(key string) []string {
	value, _ := c.GetQueryArray(key)
	return value
}

func (c *Context) getQueryCache() {
	if c.queryCache == nil {
		c.queryCache = make(url.Values)
	}
	c.queryCache, _ = url.ParseQuery(c.Request.URL.RawQuery)
}

func (c *Context) GetQueryArray(key string) ([]string, bool) {
	c.getQueryCache()
	if values, ok := c.queryCache[key]; ok && len(values) > 0 {
		return values, true
	}
	return []string{}, false
}

func (c *Context) GetQuery(key string) (string, bool) {
	if value, ok := c.GetQueryArray(key); ok {
		return value[0], ok
	}
	return "", false
}

func (c *Context) QueryMap(key string) map[string]string {
	dists, _ := c.GetQueryMap(key)
	return dists
}

func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	c.getQueryCache()
	return c.get(c.queryCache, key)
}

func (c *Context) PostForm(key string) string {
	value, _ := c.GetPostForm(key)
	return value
}

func (c *Context) DefaultPostForm(key string, defaultValue string) string {
	if value, ok := c.GetPostForm(key); ok {
		return value
	}
	return defaultValue
}

func (c *Context) GetPostForm(key string) (string, bool) {
	if values, ok := c.GetPostFormArray(key); ok {
		return values[0], ok
	}
	return "", false
}

func (c *Context) PostFormArray(key string) []string {
	values, _ := c.GetPostFormArray(key)
	return values
}

func (c *Context) getFormCache() {
	if c.formCache == nil {
		c.formCache = make(url.Values)
		req := c.Request
		if err := req.ParseMultipartForm(c.engine.MaxMultipartMemory); err != nil {
			if err != http.ErrNotMultipart {
				debugPrint("error on parse multipart form array: %v", err)
			}
		}
		c.formCache = req.PostForm
	}
}

func (c *Context) GetPostFormArray(key string) ([]string, bool) {
	c.getFormCache()
	if values := c.formCache[key]; len(values) > 0 {
		return values, true
	}
	return []string{}, false
}

func (c *Context) PostFormMap(key string) map[string]string {
	dists, _ := c.GetPostFormMap(key)
	return dists
}

func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	c.getFormCache()
	return c.get(c.formCache, key)
}

func (c *Context) get(m map[string][]string, key string) (map[string]string, bool) {
	dists := make(map[string]string)
	exist := false
	for k, v := range m {
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == "key" {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				exist = true
				dists[k[i+1:][:j]] = v[0]
			}
		}
	}
	return dists, exist
}

func (c *Context) File(filepath string) {
	http.ServeFile(c.Writer, c.Request, filepath)
}
