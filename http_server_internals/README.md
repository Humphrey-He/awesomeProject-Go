# HTTP Server Internals / HTTP 服务器内部机制

## 概述
该子项目实现一个最小化的 HTTP/1.1 请求解析与响应构建流程，用于直观理解服务端处理链路（accept -> parse -> handle -> respond）。

## Overview
This subproject implements a minimal HTTP/1.1 request parser and response writer to make the server pipeline explicit (accept -> parse -> handle -> respond).

## 目标
- 理解请求行与请求头解析
- 理解 Content-Length 驱动的 Body 读取
- 了解响应报文构建与连接处理

## Goals
- Understand request line and header parsing
- See how Content-Length drives body reads
- Understand response building and connection handling

## 设计要点
- `ParseRequest` 读取请求行、请求头和可选 Body
- `BuildResponse` 构造状态行、响应头和响应体
- `HandleConn` 串联解析 + 处理 + 写回

## Design Highlights
- `ParseRequest` reads request line, headers, and optional body
- `BuildResponse` constructs status line, headers, and body
- `HandleConn` wires parse + handler + write for one request

## 示例
```go
handler := func(r *Request) *Response {
	return &Response{StatusCode: 200, Body: []byte("pong")}
}
```

## Example
```go
handler := func(r *Request) *Response {
	return &Response{StatusCode: 200, Body: []byte("pong")}
}
```

## 测试
```bash
go test ./http_server_internals -v
```

## Tests
```bash
go test ./http_server_internals -v
```

## 扩展方向
- 支持 chunked 传输编码
- 实现 keep-alive 的请求循环与超时
- 加入简单路由与中间件链

## Extension Ideas
- Support chunked transfer encoding
- Add keep-alive loop with timeouts
- Add a tiny router and middleware chain
