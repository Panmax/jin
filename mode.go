package jin

import (
	"io"
	"os"
)

const EnvJinMode = "JIN_MODE"

const (
	// DebugMode 表明 jin 是 debug 模式
	DebugMode = "debug"

	// ReleaseMode 表明 jin 是 release 模式
	ReleaseMode = "release"

	// TestMode 表明 jin 是 test 模式
	TestMode = "test"
)

const (
	debugCode = iota
	releaseCode
	testCode
)

var DefaultWriter io.Writer = os.Stdout

var DefaultErrorWriter io.Writer = os.Stderr

var jinMode = debugCode
var modeName = DebugMode

func init() {
	mode := os.Getenv(EnvJinMode)
	SetMode(mode)
}

// SetMode 根据输入的字符串设置 jin 模式
func SetMode(value string) {
	switch value {
	case DebugMode, "":
		jinMode = debugCode
	case ReleaseMode:
		jinMode = releaseCode
	case TestMode:
		jinMode = testCode
	default:
		panic("jin mode unknown: " + value)
	}
	if value == "" {
		value = DebugMode
	}
	modeName = value
}

// Mode 返回当前 jin 模式
func Mode() string {
	return modeName
}
