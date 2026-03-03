# Go Testify 测试框架

## 概述

本项目讲解 Go 测试框架 Testify 的使用，包括断言、Suite、Mock 等核心功能。

## 核心内容

### 1. 断言库 (assert/require)

- **assert**: 失败后继续执行
- **require**: 失败后立即停止

常用断言：
```go
assert.Equal(t, expected, actual)
assert.Nil(t, err)
assert.NoError(t, err)
assert.Contains(t, container, element)
```

### 2. Suite 测试套件

```go
type UserSuite struct {
    suite.Suite
    service *UserService
}

func (s *UserSuite) SetupTest() { ... }
func (s *UserSuite) TestAddUser() { ... }
```

### 3. Mock / Stub

- Mock 对象
- Stub 实现
- 依赖注入测试

### 4. 表驱动测试

```go
tests := []struct {
    name     string
    input    int
    expected int
}{
    {"case1", 1, 2},
    {"case2", 2, 4},
}
```

## 安装

```bash
go get github.com/stretchr/testify
```

## 断言方法

| 类型 | 方法 |
|------|------|
| 基础 | Equal, NotEqual, Nil, NotNil, True, False |
| 比较 | Greater, Less, InDelta, InEpsilon |
| 容器 | Contains, Len, Empty |
| 错误 | NoError, Error, ErrorIs, ErrorContains |

## 关联项目

- [testing_practices](../testing_practices) - Go 原生测试
- [mock_practices](../mock_practices) - Mock 实践
