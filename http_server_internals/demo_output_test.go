package http_server_internals

import (
	"bufio"
	"strings"
	"testing"
)

func TestDemoOutput(t *testing.T) {
	raw := "GET /status HTTP/1.1\r\nHost: demo\r\n\r\n"
	req, err := ParseRequest(bufio.NewReader(strings.NewReader(raw)), 1024)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	resp := &Response{StatusCode: 200, Body: []byte("ok"), Close: true}
	out := string(BuildResponse(resp))
	t.Logf("method=%s path=%s", req.Method, req.Path)
	t.Logf("response=\n%s", out)
}
