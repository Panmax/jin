package jin

import (
	"net/http"
	"os"
)

type onlyFilesFs struct {
	fs http.FileSystem
}

type neuteredReaddirFile struct {
	http.File
}

// Dir 返回一个能够被 http.FileServer() 使用的 http.Filesystem。
// 它在内部被用于 router.Static()。
// 如果 listDirectory == true，它的功能和 http.Dir() 相同
//否则它返回一个阻止列出目录文件的文件系统
func Dir(root string, listDirectory bool) http.FileSystem {
	fs := http.Dir(root)
	if listDirectory {
		return fs
	}
	return &onlyFilesFs{fs}
}

func (fs onlyFilesFs) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return neuteredReaddirFile{f}, nil
}

// Readdir 重写 http.File 的默认实现
func (f neuteredReaddirFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}
