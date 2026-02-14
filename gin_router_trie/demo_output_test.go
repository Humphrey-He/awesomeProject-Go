package gin_router_trie

import "testing"

func TestDemoOutput(t *testing.T) {
	r := NewRouter()
	_ = r.Add("GET", "/users/new", func(*Context) string { return "users-new" })
	_ = r.Add("GET", "/users/:id", func(c *Context) string { return "users-id=" + c.Params["id"] })
	_ = r.Add("GET", "/assets/*filepath", func(c *Context) string { return "assets=" + c.Params["filepath"] })

	out, _ := r.Dispatch("GET", "/users/new")
	t.Logf("GET /users/new => %s", out)
	out, _ = r.Dispatch("GET", "/users/42")
	t.Logf("GET /users/42 => %s", out)
	out, _ = r.Dispatch("GET", "/assets/js/app.js")
	t.Logf("GET /assets/js/app.js => %s", out)
}


