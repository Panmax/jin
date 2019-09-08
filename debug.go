package jin

import (
	"fmt"
	"strings"
)

// 如果框架运行在 debug 模式，IsDebugging 返回 true
// 使用 SetMode(jin.ReleaseMode) 来禁用 debug 模式
func IsDebugging() bool {
	return jinMode == debugCode
}

// DebugPrintRouterFunc 指出 debug 日志输出的格式
var DebugPrintRouterFunc func(httpMethod, absolutePath, handlerName string, nuHandlers int)

func debugPrintRoute(httpMethod, absolutePath string, handlers HandlerChain) {
	if IsDebugging() {
		nuHandlers := len(handlers)
		handlerName := nameOfFunction(handlers.Last())
		if DebugPrintRouterFunc == nil {
			debugPrint("%-6s %-25s --> %s (%d handlers)\n", httpMethod, absolutePath, handlerName, nuHandlers)
		} else {
			DebugPrintRouterFunc(httpMethod, absolutePath, handlerName, nuHandlers)
		}
	}
}

func debugPrint(format string, values ...interface{}) {
	if IsDebugging() {
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		fmt.Printf(format, values)
	}
}

func debugPrintError(err error) {
	if err != nil {
		if IsDebugging() {
			fmt.Fprintf(DefaultErrorWriter, "[JIN-debug] [ERROR] %v\n", err)
		}
	}
}
