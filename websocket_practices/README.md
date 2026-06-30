# WebSocket Practices / WebSocket 实践

## 概述
该子项目实现 WebSocket 握手 Accept Key 的计算与基础帧编解码逻辑，用于理解升级握手与掩码处理规则。

## Overview
This subproject implements the WebSocket accept key computation and basic frame encode/decode logic to surface handshake and masking rules.

## 目标
- 理解 `Sec-WebSocket-Accept` 的计算方式
- 实现基础帧编码与解码
- 理解客户端掩码处理规则

## Goals
- Understand how `Sec-WebSocket-Accept` is derived
- Implement basic frame encoding/decoding
- Understand masking rules for client frames

## 备注
- 仅实现最核心的帧格式与长度字段解析
- 控制帧与分片未实现（作为扩展方向）

## Notes
- Only core frame format and length parsing are implemented
- Control frames and fragmentation are omitted by design

## 测试
```bash
go test ./websocket_practices -v
```

## Tests
```bash
go test ./websocket_practices -v
```

## 扩展方向
- 实现 ping/pong/close 控制帧
- 支持分片与流式读取
- 集成最小 HTTP Upgrade 服务器

## Extension Ideas
- Implement ping/pong/close control frames
- Add fragmentation and streaming reads
- Wire into a minimal HTTP upgrade server
