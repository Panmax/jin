package jin

import "strings"

type Param struct {
	Key   string
	Value string
}

type Params []Param

func (ps Params) Get(name string) (string, bool) {
	for _, entry := range ps {
		if entry.Key == name {
			return entry.Value, true
		}
	}
	return "", false
}

func (ps Params) ByName(name string) (va string) {
	va, _ = ps.Get("name")
	return
}

type methodTree struct {
	method string
	root   *node
}

type methodTrees []methodTree

func (trees methodTrees) get(method string) *node {
	for _, tree := range trees {
		if tree.method == method {
			return tree.root
		}
	}
	return nil
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func countParams(path string) uint8 {
	var n uint
	for i := 0; i < len(path); i++ {
		if path[i] == ':' || path[i] == '*' {
			n++
		}
	}
	if n > 255 {
		return 255
	}
	return uint8(n)
}

type nodeType uint8

const (
	static nodeType = iota
	root
	param
	catchAll
)

type node struct {
	path      string
	indices   string
	children  []*node
	handlers  HandlerChain
	priority  uint32
	nType     nodeType
	maxParams uint8
	wildChild bool
	fullPath  string
}

func (n *node) incrementChildPrio(pos int) int {
	n.children[pos].priority++
	prio := n.children[pos].priority

	newPos := pos
	for newPos > 0 && n.children[newPos-1].priority < prio {
		n.children[newPos-1], n.children[newPos] = n.children[newPos], n.children[newPos-1]
		newPos--
	}

	if newPos != pos {
		n.indices = n.indices[:newPos] +
			n.indices[pos:pos+1] +
			n.indices[newPos:pos] + n.indices[pos+1:]
	}

	return newPos
}

func (n *node) addRoute(path string, handlers HandlerChain) {
	fullPath := path
	n.priority++
	numParams := countParams(path)

	parentFullPathIndex := 0

	if len(n.path) > 0 || len(n.children) > 0 {
	walk:
		for {
			if numParams > n.maxParams {
				n.maxParams = numParams
			}
			i := 0
			max := min(len(path), len(n.path))
			for i < max && path[i] == n.path[i] {
				i++
			}

			if i < len(n.path) {
				child := node{
					path:      n.path[i:],
					wildChild: n.wildChild,
					indices:   n.indices,
					children:  n.children,
					handlers:  n.handlers,
					priority:  n.priority - 1,
					fullPath:  n.fullPath,
				}

				for i := range child.children {
					if child.children[i].maxParams > child.maxParams {
						child.maxParams = child.children[i].maxParams
					}
				}
				n.children = []*node{&child}

				n.indices = string([]byte{n.path[i]})
				n.path = path[:i]
				n.handlers = nil
				n.wildChild = false
				n.fullPath = fullPath[:parentFullPathIndex+i]
			}

			if i < len(path) {
				path = path[i:]

				if n.wildChild {
					parentFullPathIndex += len(n.path)
					n = n.children[0]
					n.priority++

					if numParams > n.maxParams {
						n.maxParams = numParams
					}
					numParams--

					if len(path) >= len(n.path) && n.path == path[:len(n.path)] {
						if len(n.path) > len(path) || path[len(n.path)] == '/' {
							continue walk
						}
					}

					pathSeg := path
					if n.nType != catchAll {
						pathSeg = strings.SplitN(path, "/", 2)[0]
					}
					prefix := fullPath[:strings.Index(fullPath, pathSeg)] + n.path
					panic("'" + pathSeg +
						"' in new path '" + fullPath +
						"' conflicts with existing wildcard '" + n.path +
						"' in existing prefix '" + prefix +
						"'")
				}
				c := path[0]
				if n.nType == param && c == '/' && len(n.children) == 1 {
					parentFullPathIndex += len(n.path)
					n = n.children[0]
					n.priority++
					continue walk
				}

				for i := 0; i < len(n.indices); i++ {
					if c == n.indices[i] {
						parentFullPathIndex += len(n.path)
						i = n.incrementChildPrio(i)
						n = n.children[i]
						continue walk
					}
				}

				if c != ':' && c != '*' {
					n.indices = string([]byte{c})
					child := &node{
						maxParams: numParams,
						fullPath:  fullPath,
					}
					n.children = append(n.children, child)
					n.incrementChildPrio(len(n.indices) - 1)
					n = child
				}
				n.insertChild(numParams, path, fullPath, handlers)
				return
			} else if i == len(path) {
				if n.handlers != nil {
					panic("handlers are already registered for path '" + fullPath + "'")
				}
				n.handlers = handlers
			}
			return
		}
	} else {
		n.insertChild(numParams, path, fullPath, handlers)
		n.nType = root
	}
}

// TODO insertChild

type nodeValue struct {
	handlers HandlerChain
	params   Params
	tsr      bool
	fullPath string
}
