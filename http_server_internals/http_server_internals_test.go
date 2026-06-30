package http_server_internals

import (
	"bufio"
	"net"
	"strings"
	"testing"
)

func TestParseRequest(t *testing.T) {
	raw := "GET /hello HTTP/1.1\r\nHost: example.com\r\nContent-Length: 5\r\n\r\nhello"
	r := bufio.NewReader(strings.NewReader(raw))
	req, err := ParseRequest(r, 1024)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if req.Method != "GET" || req.Path != "/hello" || req.Proto != "HTTP/1.1" {
		t.Fatalf("unexpected request line: %+v", req)
	}
	if req.Headers["host"] != "example.com" {
		t.Fatalf("missing host header")
	}
	if string(req.Body) != "hello" {
		t.Fatalf("body mismatch: %q", string(req.Body))
	}
}

func TestBuildResponse(t *testing.T) {
	resp := &Response{StatusCode: 200, Body: []byte("ok")}
	b := string(BuildResponse(resp))
	if !strings.HasPrefix(b, "HTTP/1.1 200 OK") {
		t.Fatalf("status line mismatch: %s", b)
	}
	if !strings.Contains(b, "Content-Length: 2") {
		t.Fatalf("missing content-length")
	}
}

func TestHandleConn(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	go func() {
		_ = HandleConn(server, func(r *Request) *Response {
			return &Response{StatusCode: 200, Body: []byte("pong")}
		}, 1024)
	}()

	_, _ = client.Write([]byte("GET /ping HTTP/1.1\r\nHost: local\r\n\r\n"))
	buf := make([]byte, 128)
	n, err := client.Read(buf)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	out := string(buf[:n])
	if !strings.Contains(out, "pong") {
		t.Fatalf("unexpected response: %s", out)
	}
}
