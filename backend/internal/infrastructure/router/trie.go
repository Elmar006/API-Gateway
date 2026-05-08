// Package router implements a tiny method+path trie router with support for
// {param} placeholders and a trailing `*` wildcard. It is purpose-built for
// the gateway: routes can be registered/replaced at runtime and the highest
// priority match wins.
package router

import (
	"strings"
	"sync"
)

// Match is the result of a Lookup hit.
type Match struct {
	Value  any               // user-supplied payload (e.g. *entity.Route)
	Params map[string]string // extracted path parameters
}

// Router is a goroutine-safe HTTP router supporting:
//   - exact segments               /api/users
//   - parameterised segments       /api/users/{id}
//   - greedy trailing wildcard     /static/*
//   - method specialisations + ANY method ("ALL").
//
// Patterns must start with "/". Trailing slashes are normalised away.
type Router struct {
	mu    sync.RWMutex
	roots map[string]*node // method -> root
}

type node struct {
	children   map[string]*node
	param      *node  // single {name}
	paramName  string // current segment name
	hasValue   bool
	value      any
	priority   int  // higher wins
	isParam    bool // current node represents a {param}
	isWildcard bool
}

// New returns an empty router.
func New() *Router {
	return &Router{roots: map[string]*node{}}
}

// Add registers a value for the (method, pattern) tuple. If method is empty or
// "ALL", the route matches any method. priority breaks ties: the highest
// priority wins. Re-adding an identical (method, pattern) replaces the value.
func (r *Router) Add(method, pattern string, priority int, value any) {
	method = normalizeMethod(method)
	pattern = normalizePattern(pattern)

	r.mu.Lock()
	defer r.mu.Unlock()

	root, ok := r.roots[method]
	if !ok {
		root = &node{children: map[string]*node{}}
		r.roots[method] = root
	}

	cur := root
	if pattern == "/" {
		cur.setValue(value, priority)
		return
	}
	segments := strings.Split(strings.TrimPrefix(pattern, "/"), "/")
	for i, seg := range segments {
		switch {
		case seg == "*" || (i == len(segments)-1 && seg != "" && strings.HasSuffix(seg, "*")):
			// trailing wildcard – everything below is consumed
			if cur.children == nil {
				cur.children = map[string]*node{}
			}
			child, ok := cur.children["*"]
			if !ok {
				child = &node{isWildcard: true}
				cur.children["*"] = child
			}
			child.setValue(value, priority)
			return
		case strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}"):
			name := strings.TrimSuffix(strings.TrimPrefix(seg, "{"), "}")
			if cur.param == nil {
				cur.param = &node{children: map[string]*node{}, isParam: true}
			}
			cur.param.paramName = name
			cur = cur.param
		default:
			if cur.children == nil {
				cur.children = map[string]*node{}
			}
			child, ok := cur.children[seg]
			if !ok {
				child = &node{children: map[string]*node{}}
				cur.children[seg] = child
			}
			cur = child
		}
	}
	cur.setValue(value, priority)
}

// Lookup returns the registered value matching (method, path), or ok=false.
// If no method-specific match is found, the "ALL" tree is consulted.
func (r *Router) Lookup(method, path string) (Match, bool) {
	method = normalizeMethod(method)
	path = normalizePattern(path)

	r.mu.RLock()
	defer r.mu.RUnlock()

	if root, ok := r.roots[method]; ok {
		if m, found := walk(root, splitSegs(path), nil); found {
			return m, true
		}
	}
	if method != "ALL" {
		if root, ok := r.roots["ALL"]; ok {
			return walk(root, splitSegs(path), nil)
		}
	}
	return Match{}, false
}

// Reset clears all registered routes.
func (r *Router) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.roots = map[string]*node{}
}

func walk(n *node, segs []string, params map[string]string) (Match, bool) {
	if len(segs) == 0 {
		if n.hasValue {
			return Match{Value: n.value, Params: cloneParams(params)}, true
		}
		// allow trailing wildcard match for "/foo" against "/foo/*"
		if w, ok := n.children["*"]; ok && w.hasValue {
			return Match{Value: w.value, Params: cloneParams(params)}, true
		}
		return Match{}, false
	}
	seg, rest := segs[0], segs[1:]

	// 1. Exact segment match
	if child, ok := n.children[seg]; ok {
		if m, found := walk(child, rest, params); found {
			return m, true
		}
	}
	// 2. Parameter match.
	//
	// Allocate a *fresh* map for this branch so that if the recursive walk
	// fails and we fall through to the wildcard at step 3, the caller's
	// `params` map is not polluted with the failed attempt's key/value.
	// Without this, given routes /api/{version}/users and /api/* a request
	// for /api/v2/docs would: try {version}=v2 (mutates params), backtrack,
	// match the wildcard, and return params={"version":"v2"} — phantom
	// parameters that are forwarded to the upstream as X-Path-Param-* headers.
	if n.param != nil {
		next := make(map[string]string, len(params)+1)
		for k, v := range params {
			next[k] = v
		}
		next[n.param.paramName] = seg
		if m, found := walk(n.param, rest, next); found {
			return m, true
		}
	}
	// 3. Wildcard match (consumes the rest)
	if w, ok := n.children["*"]; ok && w.hasValue {
		return Match{Value: w.value, Params: cloneParams(params)}, true
	}
	return Match{}, false
}

func (n *node) setValue(v any, prio int) {
	if !n.hasValue || prio >= n.priority {
		n.hasValue = true
		n.value = v
		n.priority = prio
	}
}

func cloneParams(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func splitSegs(path string) []string {
	if path == "" || path == "/" {
		return nil
	}
	return strings.Split(strings.TrimPrefix(path, "/"), "/")
}

func normalizePattern(p string) string {
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if len(p) > 1 && strings.HasSuffix(p, "/") {
		p = strings.TrimRight(p, "/")
	}
	return p
}

func normalizeMethod(m string) string {
	m = strings.TrimSpace(strings.ToUpper(m))
	if m == "" {
		return "ALL"
	}
	return m
}
