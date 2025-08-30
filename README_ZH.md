[![Go Doc](https://img.shields.io/badge/api-reference-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/icloudza/fxjson?utm_source=godoc)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.18-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Performance](https://img.shields.io/badge/performance-fast-orange.svg)](#性能对比)

[📄 English Documentation](README.md)

# FxJSON 🔥 - 高性能Go JSON解析库

FxJSON 是一个专注性能的Go JSON解析库，提供高效的JSON遍历和访问能力。相比标准库有显著的性能提升，同时保持内存安全和易用性。

## 🚀 核心特性

- **🔥 高性能**: 数组遍历快67倍，对象遍历快20倍
- **⚡ 内存高效**: 核心操作零内存分配
- **🛡️ 内存安全**: 完备的边界检查和安全机制
- **🎯 易于使用**: 链式调用，直观的API设计
- **🔧 功能完整**: 高级查询、数据验证、缓存等功能

## 📦 安装

```bash
go get github.com/icloudza/fxjson
```

## 🎯 快速开始

```go
package main

import (
    "fmt"
    "github.com/icloudza/fxjson"
)

func main() {
    jsonData := []byte(`{
        "name": "Alice",
        "age": 30,
        "profile": {
            "city": "北京",
            "hobby": "coding"
        }
    }`)

    node := fxjson.FromBytes(jsonData)

    // 安全访问，使用默认值
    name := node.Get("name").StringOr("未知")
    age := node.Get("age").IntOr(0)
    city := node.GetPath("profile.city").StringOr("")

    fmt.Printf("姓名: %s, 年龄: %d, 城市: %s\n", name, age, city)
}
```

## 📊 性能亮点

| 操作        | FxJSON   | 标准库      | 性能提升      |
|-----------|----------|----------|-----------|
| 数组遍历      | 30.27 ns | 2044 ns  | **快67倍** |
| 对象遍历      | 104.7 ns | 2115 ns  | **快20倍** |
| 缓存解析      | 1,396 ns | 5,542 ns | **快4倍**  |

## 🚀 核心特性

### 零分配遍历
```go
// 比标准库快67倍
users.ArrayForEach(func(index int, user fxjson.Node) bool {
    name := user.Get("name").StringOr("")
    fmt.Printf("用户 %d: %s\n", index+1, name)
    return true
})
```

### 安全的默认值
```go
// 无需错误处理
name := node.Get("name").StringOr("未知")
age := node.Get("age").IntOr(0)
active := node.Get("active").BoolOr(false)
```

### 内置验证
```go
if node.Get("email").IsValidEmail() {
    fmt.Println("✅ 邮箱格式正确")
}
```

### 高级功能
- SQL风格的查询和过滤
- 数据聚合和统计
- 高性能缓存
- 结构体编码/解码
- 嵌套JSON展开
- 批量操作

## 📚 完整文档

查看完整的教程、高级示例和详细API参考，请访问我们的wiki：

**🔗 [完整文档Wiki](https://github.com/icloudza/fxjson/wiki)**

Wiki包含：
- 详细教程和示例
- 高级功能和最佳实践
- 实际使用场景
- 性能优化指南
- 完整API参考

## 🤝 贡献

欢迎提交Issue和Pull Request！

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

**FxJSON - 让JSON解析飞起来！** 🚀

---

## 📊 完整基准测试结果

### 核心操作性能

| 操作 | FxJSON | 标准库 | 性能提升 | 内存优势 |
|------|--------|--------|----------|----------|
| Get | 24.88 ns | 2012 ns | **快80.8倍** | 0 vs 1984 B |
| GetPath | 111.5 ns | 2055 ns | **快18.4倍** | 0 vs 1984 B |
| Int转换 | 16.70 ns | 2026 ns | **快121.3倍** | 0 vs 1984 B |
| Float转换 | 7.688 ns | 2051 ns | **快266.7倍** | 0 vs 1984 B |
| Bool转换 | 3.684 ns | 2149 ns | **快583.2倍** | 0 vs 1984 B |
| String访问 | 5.402 ns | 2083 ns | **快385.6倍** | 0 vs 1984 B |
| 数组长度 | 20.70 ns | 2152 ns | **快103.9倍** | 0 vs 1984 B |
| 数组索引 | 18.42 ns | 2134 ns | **快115.9倍** | 0 vs 1984 B |
| 键存在检查 | 0.2454 ns | 2110 ns | **快8598倍** | 0 vs 1984 B |

### 遍历操作

| 操作 | FxJSON | 标准库 | 性能提升 | 内存优势 |
|------|--------|--------|----------|----------|
| 对象遍历 | 108.9 ns | 2142 ns | **快19.7倍** | 0 vs 1984 B |
| 数组遍历 | 30.21 ns | 2119 ns | **快70.2倍** | 0 vs 1984 B |
| 深度遍历 | 1536 ns | 2891 ns | **快1.9倍** | 3056 vs 2289 B |
| 复杂遍历 | 1310 ns | 3505 ns | **快2.7倍** | 0 vs 4136 B |
| 大数据遍历 | 12.8 µs | 17.4 µs | **快1.4倍** | 19136 vs 14698 B |

### 结构体操作

| 操作 | FxJSON | 标准库 | 性能提升 | 内存优势 |
|------|--------|--------|----------|----------|
| 基础解码 | 967.8 ns | 1877 ns | **快1.9倍** | 256 vs 736 B |
| DecodeStruct | 939.5 ns | - | - | 256 B |
| DecodeStructFast | 868.6 ns | - | - | 256 B |
| 复杂解码 | 2668 ns | 3355 ns | **快1.3倍** | 592 vs 1520 B |
| 大数据解码 | 9.53 µs | 11.8 µs | **快1.2倍** | 1864 vs 4640 B |

### 高级功能性能

| 功能特性 | 操作耗时 | 内存使用 | 分配次数 | 说明 |
|----------|----------|----------|----------|------|
| 基础解析 | 5,290 ns | 6,448 B | 45 allocs | 标准JSON解析 |
| **缓存解析** | **641.8 ns** | **20 B** | **2 allocs** | **快8.2倍，内存减少99.7%** |
| 简单查询 | 3,386 ns | 640 B | 14 allocs | 基础过滤 |
| 复杂查询 | 4,986 ns | 1,720 B | 52 allocs | 多条件查询和排序 |
| 数据聚合 | 4,804 ns | 2,640 B | 32 allocs | 统计计算 |
| 数据变换 | 478.7 ns | 368 B | 5 allocs | 字段映射和类型转换 |
| 数据验证 | 216.6 ns | 360 B | 4 allocs | 基于规则的验证 |
| 流式处理 | 3,250 ns | 0 B | 0 allocs | 零分配流式数据处理 |
| 大数据查询 | 1.28 ms | 80 B | 2 allocs | 100条记录处理 |
| JSON差异对比 | 18.2 µs | 2,787 B | 197 allocs | 变更检测 |
| 空字符串处理 | 2,777 ns | 1,664 B | 27 allocs | 安全的空字符串处理 |

### 序列化性能

| 操作 | 时间 | 内存 | 分配次数 | 说明 |
|------|------|------|----------|------|
| Marshal | 652.1 ns | 424 B | 9 allocs | 标准序列化 |
| FastMarshal | 226.7 ns | 136 B | 2 allocs | 高性能序列化 |
| StructMarshal | 267.1 ns | 136 B | 2 allocs | 直接结构体序列化 |

### 默认值函数

| 函数 | 时间 | 内存 | 分配次数 |
|------|------|------|----------|
| StringOr | 23.56 ns | 0 B | 0 allocs |
| IntOr | 28.34 ns | 0 B | 0 allocs |
| FloatOr | 40.89 ns | 0 B | 0 allocs |

*基准测试结果基于 Apple M4 Pro，Go 1.24.6*