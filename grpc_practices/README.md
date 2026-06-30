# gRPC Practices (Framing + Interceptors) / gRPC 实践（帧格式与拦截器）

## 概述
该子项目模拟 gRPC 的消息帧格式与一元拦截器链，不依赖 protobuf 代码生成，重点理解字节传输与横切逻辑的封装顺序。

## Overview
This subproject models gRPC wire framing and a unary interceptor chain without protobuf code generation, focusing on bytes-on-the-wire and cross-cutting concerns.

## 目标
- 理解 gRPC 消息帧格式（1 字节压缩标志 + 4 字节长度 + payload）
- 构建一元拦截器链的执行顺序
- 通过 context 传递元数据

## Goals
- Understand gRPC framing (1-byte flag + 4-byte length + payload)
- Build a unary interceptor chain
- Pass metadata through context

## 备注
- 这是 gRPC 的概念化切片，不包含 HTTP/2 传输与 protobuf 编解码
- 帧格式与官方定义保持一致

## Notes
- This is a conceptual slice of gRPC; no HTTP/2 transport or protobuf codecs
- The framer layout matches the official gRPC length-prefixed message format

## 测试
```bash
go test ./grpc_practices -v
```

## Tests
```bash
go test ./grpc_practices -v
```

## 扩展方向
- 接入 HTTP/2 以验证流式读写
- 增加压缩协商与可插拔编解码
- 增加 deadline 传播与取消链路

## Extension Ideas
- Plug the framer into HTTP/2 and test streaming
- Add compression negotiation and pluggable codecs
- Implement deadline propagation and cancellation tests
