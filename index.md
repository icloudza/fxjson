---
layout: home

hero:
  name: "FxJSON"
  text: "高性能 Go JSON 解析库"
  tagline: 专为现代 Go 应用设计的零分配、高性能 JSON 处理库。数组遍历快 67 倍，对象遍历快 20 倍。
  actions:
    - theme: brand
      text: 5分钟快速上手
      link: /guide/quick-start
    - theme: alt
      text: 完整 API 文档
      link: /api/
    - theme: alt
      text: 查看源码
      link: https://github.com/icloudza/fxjson

features:
  - title: 极致性能优化
    details: 采用零分配设计，数组遍历比标准库快 67 倍，对象遍历快 20 倍。内置智能缓存系统，重复访问场景性能提升 4 倍。
  - title: 开发者友好
    details: 提供直观的链式 API，支持安全的默认值处理。无需复杂的错误处理代码，让您专注于业务逻辑。
  - title: 功能完整
    details: 支持路径访问、类型转换、数据验证、结构体编解码、批量操作等全套 JSON 处理功能，一个库解决所有需求。
  - title: 生产就绪
    details: 内存安全保证，完备的边界检查，零外部依赖。经过严格测试，可直接用于生产环境。
  - title: 简单集成
    details: 一条命令安装，API 设计与 gjson 兼容，可无缝迁移现有项目。支持 Go 1.18+ 所有版本。
  - title: 性能透明
    details: 提供详细的基准测试报告，每个操作的性能数据清晰可见，帮助您做出明智的技术选择。
---

## 为什么选择 FxJSON？

### 传统 JSON 处理的痛点
- 标准库 `encoding/json` 性能不足，大量内存分配
- 第三方库要么功能有限，要么学习成本高
- 错误处理复杂，代码冗余

### FxJSON 的解决方案
- **零分配核心操作**：关键路径无内存分配，GC 压力小
- **智能缓存机制**：重复访问自动优化，性能持续提升  
- **安全默认值**：`StringOr()` `IntOr()` 等方法优雅处理异常
- **完整功能集**：从基础解析到高级查询，一站式解决

## 快速体验

```go
package main

import (
    "fmt"
    "github.com/icloudza/fxjson"
)

func main() {
    // 1. 解析 JSON（零分配）
    json := `{"user":{"name":"张三","age":25,"skills":["Go","JSON"]}}`
    node := fxjson.FromString(json)
    
    // 2. 安全访问（自动处理错误）
    name := node.GetPath("user.name").StringOr("未知")
    age := node.GetPath("user.age").IntOr(0)
    
    // 3. 高性能遍历（比标准库快67倍）
    node.GetPath("user.skills").ArrayForEach(func(i int, skill fxjson.Node) bool {
        fmt.Printf("技能 %d: %s\n", i+1, skill.StringOr(""))
        return true
    })
    
    fmt.Printf("%s 今年 %d 岁\n", name, age)
}
```

## 性能对比一览

| 操作类型 | FxJSON | 标准库 | 性能提升 | 内存优势 |
|---------|--------|--------|----------|----------|
| 数组遍历 | 30ns | 2044ns | **67.5x** | 0 vs 1984B |
| 对象遍历 | 104ns | 2115ns | **20.2x** | 0 vs 1984B |
| 字段访问 | 25ns | 2012ns | **80.8x** | 0 vs 1984B |
| 缓存访问 | 642ns | 5542ns | **8.6x** | 20B vs 6448B |

*基准测试环境：Apple M4 Pro, Go 1.24*

## 适用场景

- **API 服务**：处理大量 JSON 请求响应
- **配置文件解析**：应用启动时的配置读取
- **数据处理**：ETL 流程中的 JSON 数据转换
- **微服务通信**：服务间 JSON 格式数据交换
- **日志分析**：结构化日志数据的解析处理

---

