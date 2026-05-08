package router

import "testing"

func TestRouterMatching(t *testing.T) {
	r := New()
	r.Add("GET", "/users", 0, "list-users")
	r.Add("GET", "/users/{id}", 0, "get-user")
	r.Add("POST", "/users", 0, "create-user")
	r.Add("ALL", "/static/*", 0, "static")
	r.Add("GET", "/api/{ver}/items/{id}", 0, "deep")
	r.Add("GET", "/health", 5, "health-high")
	r.Add("GET", "/health", 1, "health-low")

	tests := []struct {
		name       string
		method     string
		path       string
		wantValue  any
		wantParams map[string]string
		wantOK     bool
	}{
		{name: "exact list", method: "GET", path: "/users", wantValue: "list-users", wantOK: true},
		{name: "param users id", method: "GET", path: "/users/42", wantValue: "get-user", wantParams: map[string]string{"id": "42"}, wantOK: true},
		{name: "POST users", method: "POST", path: "/users", wantValue: "create-user", wantOK: true},
		{name: "ALL static", method: "PUT", path: "/static/css/app.css", wantValue: "static", wantOK: true},
		{name: "deep params", method: "GET", path: "/api/v2/items/77", wantValue: "deep", wantParams: map[string]string{"ver": "v2", "id": "77"}, wantOK: true},
		{name: "priority wins", method: "GET", path: "/health", wantValue: "health-high", wantOK: true},
		{name: "no match", method: "GET", path: "/missing", wantOK: false},
		{name: "method falls through to ALL", method: "DELETE", path: "/static/x", wantValue: "static", wantOK: true},
		{name: "trailing slash normalization", method: "GET", path: "/users/", wantValue: "list-users", wantOK: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ok := r.Lookup(tt.method, tt.path)
			if ok != tt.wantOK {
				t.Fatalf("ok=%v, want %v (match=%+v)", ok, tt.wantOK, m)
			}
			if !ok {
				return
			}
			if m.Value != tt.wantValue {
				t.Errorf("value=%v, want %v", m.Value, tt.wantValue)
			}
			if len(tt.wantParams) > 0 {
				for k, v := range tt.wantParams {
					if m.Params[k] != v {
						t.Errorf("param %s=%q, want %q", k, m.Params[k], v)
					}
				}
			}
		})
	}
}

func TestRouterReset(t *testing.T) {
	r := New()
	r.Add("GET", "/x", 0, "v")
	if _, ok := r.Lookup("GET", "/x"); !ok {
		t.Fatal("expected match before reset")
	}
	r.Reset()
	if _, ok := r.Lookup("GET", "/x"); ok {
		t.Fatal("expected miss after reset")
	}
}

// Regression: when a {param} branch was tried and failed, the params map was
// previously mutated in place. Falling through to a wildcard then returned the
// stale parameter as a phantom path-param. The fix allocates a fresh map for
// the {param} branch so the caller's map is never polluted.
func TestRouter_ParamBacktrackDoesNotLeakIntoWildcard(t *testing.T) {
	r := New()
	r.Add("GET", "/api/{version}/users", 0, "versioned-users")
	r.Add("GET", "/api/*", 0, "api-wildcard")

	// /api/v2/users matches the {version} route → params should contain version.
	if m, ok := r.Lookup("GET", "/api/v2/users"); !ok {
		t.Fatalf("expected versioned-users to match /api/v2/users")
	} else {
		if m.Value != "versioned-users" {
			t.Errorf("value=%v, want versioned-users", m.Value)
		}
		if got := m.Params["version"]; got != "v2" {
			t.Errorf("params[version]=%q, want v2", got)
		}
	}

	// /api/v2/docs falls through {version} (no /users segment) → wildcard.
	// The wildcard must NOT carry version=v2 from the failed param attempt.
	m, ok := r.Lookup("GET", "/api/v2/docs")
	if !ok {
		t.Fatalf("expected api-wildcard to match /api/v2/docs")
	}
	if m.Value != "api-wildcard" {
		t.Errorf("value=%v, want api-wildcard", m.Value)
	}
	if _, has := m.Params["version"]; has {
		t.Errorf("wildcard match leaked params=%v from failed {version} branch", m.Params)
	}
	if len(m.Params) != 0 {
		t.Errorf("wildcard params=%v, want empty", m.Params)
	}
}

// Regression: nested {param} branches must not pollute each other if the
// deeper branch backtracks before consuming a segment.
func TestRouter_NestedParamBacktrackIsolated(t *testing.T) {
	r := New()
	r.Add("GET", "/a/{x}/b/{y}/c", 0, "deep")
	r.Add("GET", "/a/{x}/*", 0, "shallow-wildcard")

	// /a/foo/b/bar/d → tries deep, fails at "d" vs "c", falls back to wildcard.
	// Wildcard must see x=foo (set by outer {x} that *did* match), but must
	// NOT carry y=bar (set by inner {y} that never reached its terminal node).
	m, ok := r.Lookup("GET", "/a/foo/b/bar/d")
	if !ok {
		t.Fatalf("expected shallow-wildcard to match")
	}
	if m.Value != "shallow-wildcard" {
		t.Errorf("value=%v, want shallow-wildcard", m.Value)
	}
	if got := m.Params["x"]; got != "foo" {
		t.Errorf("params[x]=%q, want foo", got)
	}
	if _, has := m.Params["y"]; has {
		t.Errorf("phantom y=%q leaked from failed inner {y} branch", m.Params["y"])
	}
}

func TestNormalize(t *testing.T) {
	if normalizePattern("") != "/" {
		t.Errorf("empty pattern")
	}
	if normalizePattern("api/users") != "/api/users" {
		t.Errorf("missing slash: %s", normalizePattern("api/users"))
	}
	if normalizePattern("/foo/") != "/foo" {
		t.Errorf("trailing slash")
	}
	if normalizeMethod("") != "ALL" {
		t.Errorf("empty method default")
	}
	if normalizeMethod("post") != "POST" {
		t.Errorf("upper")
	}
}
