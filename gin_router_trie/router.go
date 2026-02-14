// Package gin_router_trie 实现一个类似 Gin 的高效前缀树路由匹配。
package gin_router_trie

import (
	"errors"
	"strings"
)

var (
	ErrInvalidPath      = errors.New("invalid route path")
	ErrRouteConflict    = errors.New("route conflict")
	ErrMethodNotAllowed = errors.New("method not allowed")
)

// Handler 表示业务处理函数，这里为简化返回一个字符串结果。
type Handler func(*Context) string

// Context 是一次请求的上下文，包含方法、路径以及路径参数。
type Context struct {
	Method string
	Path   string
	Params map[string]string
}

// nodeType 表示 Trie 节点类型：静态段、参数段、通配段。
type nodeType int

const (
	nodeStatic nodeType = iota
	nodeParam
	nodeWildcard
)

// node 是前缀树中的一个节点，保存一段路径以及其下级子节点。
type node struct {
	segment   string
	typ       nodeType
	paramName string
	children  []*node
	handler   Handler
	routePath string
}

// Router 为每个 HTTP Method 维护一棵独立的前缀树。
type Router struct {
	trees map[string]*node // method -> root
}

// NewRouter 创建一个新的路由树。
func NewRouter() *Router {
	return &Router{trees: make(map[string]*node)}
}

// Add 注册一条路由规则，例如 GET /users/:id。
// 为简化起见，这里不支持同一路径下混合多个 Handler。
func (r *Router) Add(method, path string, h Handler) error {
	if method == "" || !strings.HasPrefix(path, "/") || h == nil {
		return ErrInvalidPath
	}
	segs := splitPath(path)
	root := r.trees[method]
	if root == nil {
		root = &node{segment: "/", typ: nodeStatic}
		r.trees[method] = root
	}
	cur := root
	for i, seg := range segs {
		typ, paramName, err := classify(seg)
		if err != nil {
			return err
		}
		if typ == nodeWildcard && i != len(segs)-1 {
			return ErrInvalidPath
		}
		child := findChild(cur, typ, seg, paramName)
		if child == nil {
			child = &node{segment: seg, typ: typ, paramName: paramName}
			cur.children = append(cur.children, child)
		}
		cur = child
	}
	if cur.handler != nil {
		return ErrRouteConflict
	}
	cur.handler = h
	cur.routePath = path
	return nil
}

// Match 在路由树中查找匹配的 Handler，并解析路径参数。
func (r *Router) Match(method, path string) (Handler, *Context, bool) {
	root := r.trees[method]
	if root == nil {
		return nil, nil, false
	}
	segs := splitPath(path)
	ctx := &Context{Method: method, Path: path, Params: map[string]string{}}
	h, ok := matchNode(root, segs, 0, ctx)
	if !ok {
		return nil, nil, false
	}
	return h, ctx, true
}

// Dispatch 是对外统一的分发入口，用于运行测试或 Demo。
func (r *Router) Dispatch(method, path string) (string, error) {
	if _, ok := r.trees[method]; !ok {
		return "", ErrMethodNotAllowed
	}
	h, ctx, ok := r.Match(method, path)
	if !ok {
		return "", ErrNotFound
	}
	return h(ctx), nil
}

var ErrNotFound = errors.New("route not found")

func matchNode(cur *node, segs []string, idx int, ctx *Context) (Handler, bool) {
	if idx == len(segs) {
		if cur.handler != nil {
			return cur.handler, true
		}
		// allow wildcard child to match empty tail
		for _, ch := range cur.children {
			if ch.typ == nodeWildcard && ch.handler != nil {
				ctx.Params[ch.paramName] = ""
				return ch.handler, true
			}
		}
		return nil, false
	}
	seg := segs[idx]

	// 1) static
	for _, ch := range cur.children {
		if ch.typ == nodeStatic && ch.segment == seg {
			if h, ok := matchNode(ch, segs, idx+1, ctx); ok {
				return h, true
			}
		}
	}
	// 2) param
	for _, ch := range cur.children {
		if ch.typ == nodeParam {
			old, existed := ctx.Params[ch.paramName]
			ctx.Params[ch.paramName] = seg
			if h, ok := matchNode(ch, segs, idx+1, ctx); ok {
				return h, true
			}
			if existed {
				ctx.Params[ch.paramName] = old
			} else {
				delete(ctx.Params, ch.paramName)
			}
		}
	}
	// 3) wildcard
	for _, ch := range cur.children {
		if ch.typ == nodeWildcard {
			ctx.Params[ch.paramName] = strings.Join(segs[idx:], "/")
			if ch.handler != nil {
				return ch.handler, true
			}
		}
	}
	return nil, false
}

func classify(seg string) (nodeType, string, error) {
	if seg == "" {
		return nodeStatic, "", ErrInvalidPath
	}
	if strings.HasPrefix(seg, ":") {
		name := strings.TrimPrefix(seg, ":")
		if name == "" {
			return nodeStatic, "", ErrInvalidPath
		}
		return nodeParam, name, nil
	}
	if strings.HasPrefix(seg, "*") {
		name := strings.TrimPrefix(seg, "*")
		if name == "" {
			return nodeStatic, "", ErrInvalidPath
		}
		return nodeWildcard, name, nil
	}
	return nodeStatic, "", nil
}

func findChild(cur *node, typ nodeType, seg, paramName string) *node {
	for _, ch := range cur.children {
		if ch.typ != typ {
			continue
		}
		switch typ {
		case nodeStatic:
			if ch.segment == seg {
				return ch
			}
		case nodeParam, nodeWildcard:
			if ch.paramName == paramName {
				return ch
			}
		}
	}
	return nil
}

func splitPath(p string) []string {
	if p == "/" {
		return nil
	}
	raw := strings.Split(strings.Trim(p, "/"), "/")
	out := make([]string, 0, len(raw))
	for _, s := range raw {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
