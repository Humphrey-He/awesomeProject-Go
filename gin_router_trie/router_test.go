package gin_router_trie

import "testing"

func TestStaticParamWildcardPriority(t *testing.T) {
	r := NewRouter()
	_ = r.Add("GET", "/users/new", func(ctx *Context) string { return "static" })
	_ = r.Add("GET", "/users/:id", func(ctx *Context) string { return "param:" + ctx.Params["id"] })
	_ = r.Add("GET", "/users/*rest", func(ctx *Context) string { return "wild:" + ctx.Params["rest"] })

	out, err := r.Dispatch("GET", "/users/new")
	if err != nil || out != "static" {
		t.Fatalf("unexpected static out=%s err=%v", out, err)
	}
	out, err = r.Dispatch("GET", "/users/123")
	if err != nil || out != "param:123" {
		t.Fatalf("unexpected param out=%s err=%v", out, err)
	}
	out, err = r.Dispatch("GET", "/users/a/b")
	if err != nil || out != "wild:a/b" {
		t.Fatalf("unexpected wildcard out=%s err=%v", out, err)
	}
}

func TestMethodIsolation(t *testing.T) {
	r := NewRouter()
	_ = r.Add("GET", "/ping", func(*Context) string { return "pong-get" })
	_ = r.Add("POST", "/ping", func(*Context) string { return "pong-post" })
	out, _ := r.Dispatch("GET", "/ping")
	if out != "pong-get" {
		t.Fatalf("unexpected out=%s", out)
	}
	out, _ = r.Dispatch("POST", "/ping")
	if out != "pong-post" {
		t.Fatalf("unexpected out=%s", out)
	}
}

func TestRouteConflict(t *testing.T) {
	r := NewRouter()
	if err := r.Add("GET", "/a/:id", func(*Context) string { return "1" }); err != nil {
		t.Fatalf("unexpected err=%v", err)
	}
	if err := r.Add("GET", "/a/:id", func(*Context) string { return "2" }); err == nil {
		t.Fatal("expected route conflict")
	}
}


