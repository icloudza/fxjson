[![Go Doc](https://img.shields.io/badge/api-reference-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/icloudza/fxjson?utm_source=godoc)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.18-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Performance](https://img.shields.io/badge/performance-fast-orange.svg)](#performance-comparison)

[📄 中文文档 / Chinese Documentation](README_ZH.md)

# FxJSON ![Flame](flame.png) - High-Performance JSON Parser for Go

FxJSON is a Go JSON parsing library focused on performance, providing efficient JSON traversal and access capabilities.
It offers improved performance compared to the standard library while maintaining memory safety and ease of use.

## 🚀 Core Features

- **🔥 High Performance**: Up to 67x faster array traversal, 20x faster object traversal
- **⚡ Memory Efficient**: Zero-allocation core operations
- **🛡️ Memory Safety**: Proper boundary checking and safety mechanisms
- **🎯 Easy to Use**: Chainable calls with intuitive API design
- **🔧 Feature Complete**: Advanced querying, data validation, caching, and more

## 📦 Installation

```bash
go get github.com/icloudza/fxjson
```

## 🎯 Quick Start

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
            "city": "Beijing",
            "hobby": "coding"
        }
    }`)

    node := fxjson.FromBytes(jsonData)

    // Safe access with default values
    name := node.Get("name").StringOr("Unknown")
    age := node.Get("age").IntOr(0)
    city := node.GetPath("profile.city").StringOr("")

    fmt.Printf("Name: %s, Age: %d, City: %s\n", name, age, city)
}
```

## 📊 Performance Highlights

| Operation            | FxJSON   | Standard Library | Performance Gain |
|----------------------|----------|------------------|------------------|
| Array Traversal      | 30.27 ns | 2044 ns          | **67.5x faster** |
| Object Traversal     | 104.7 ns | 2115 ns          | **20.2x faster** |
| Cached Parsing       | 1,396 ns | 5,542 ns         | **4x faster**    |

## 🚀 Key Features

### Zero-Allocation Traversal
```go
// 67x faster than standard library
users.ArrayForEach(func(index int, user fxjson.Node) bool {
    name := user.Get("name").StringOr("")
    fmt.Printf("User %d: %s\n", index+1, name)
    return true
})
```

### Safe Default Values
```go
// No error handling needed
name := node.Get("name").StringOr("Unknown")
age := node.Get("age").IntOr(0)
active := node.Get("active").BoolOr(false)
```

### Built-in Validation
```go
if node.Get("email").IsValidEmail() {
    fmt.Println("✅ Valid email")
}
```

### Advanced Features
- SQL-style querying and filtering
- Data aggregation and statistics
- High-performance caching
- Struct encoding/decoding
- Nested JSON expansion
- Batch operations

## 📚 Complete Documentation

For comprehensive tutorials, advanced examples, and detailed API reference, visit our wiki:

**🔗 [Complete Documentation Wiki](https://github.com/icloudza/fxjson/wiki)**

The wiki includes:
- Detailed tutorials and examples
- Advanced features and best practices
- Real-world usage scenarios
- Performance optimization guides
- Complete API reference

## 🤝 Contributing

Issues and Pull Requests are welcome!

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details

**FxJSON - Make JSON parsing fly!** 🚀

---

## 📊 Complete Benchmark Results

### Core Operations Performance

| Operation        | FxJSON    | Standard Library | Performance Gain  | Memory Advantage |
|------------------|-----------|------------------|-------------------|------------------|
| Get              | 24.88 ns  | 2012 ns          | **80.8x faster**  | 0 vs 1984 B      |
| GetPath          | 111.5 ns  | 2055 ns          | **18.4x faster**  | 0 vs 1984 B      |
| Int Conversion   | 16.70 ns  | 2026 ns          | **121.3x faster** | 0 vs 1984 B      |
| Float Conversion | 7.688 ns  | 2051 ns          | **266.7x faster** | 0 vs 1984 B      |
| Bool Conversion  | 3.684 ns  | 2149 ns          | **583.2x faster** | 0 vs 1984 B      |
| String Access    | 5.402 ns  | 2083 ns          | **385.6x faster** | 0 vs 1984 B      |
| Array Length     | 20.70 ns  | 2152 ns          | **103.9x faster** | 0 vs 1984 B      |
| Array Index      | 18.42 ns  | 2134 ns          | **115.9x faster** | 0 vs 1984 B      |
| Key Existence    | 0.2454 ns | 2110 ns          | **8598x faster**  | 0 vs 1984 B      |

### Traversal Operations

| Operation            | FxJSON   | Standard Library | Performance Gain | Memory Advantage |
|----------------------|----------|------------------|------------------|------------------|
| Object ForEach       | 108.9 ns | 2142 ns          | **19.7x faster** | 0 vs 1984 B      |
| Array ForEach        | 30.21 ns | 2119 ns          | **70.2x faster** | 0 vs 1984 B      |
| Deep Walk            | 1536 ns  | 2891 ns          | **1.9x faster**  | 3056 vs 2289 B   |
| Complex Traversal    | 1310 ns  | 3505 ns          | **2.7x faster**  | 0 vs 4136 B      |
| Large Data Traversal | 12.8 µs  | 17.4 µs          | **1.4x faster**  | 19136 vs 14698 B |

### Struct Operations

| Operation | FxJSON | Standard Library | Performance Gain | Memory Advantage |
|-----------|--------|------------------|------------------|------------------|
| Basic Decode | 967.8 ns | 1877 ns | **1.9x faster** | 256 vs 736 B |
| DecodeStruct | 939.5 ns | - | - | 256 B |
| DecodeStructFast | 868.6 ns | - | - | 256 B |
| Complex Decode | 2668 ns | 3355 ns | **1.3x faster** | 592 vs 1520 B |
| Large Decode | 9.53 µs | 11.8 µs | **1.2x faster** | 1864 vs 4640 B |

### Advanced Features Performance

| Feature               | Operation Time | Memory Usage | Allocations  | Note                               |
|-----------------------|----------------|--------------|--------------|------------------------------------|
| Basic Parsing         | 5,290 ns       | 6,448 B      | 45 allocs    | Standard JSON parsing              |
| **Cached Parsing**    | **641.8 ns**   | **20 B**     | **2 allocs** | **8.2x faster, 99.7% less memory** |
| Simple Query          | 3,386 ns       | 640 B        | 14 allocs    | Basic filtering                    |
| Complex Query         | 4,986 ns       | 1,720 B      | 52 allocs    | Multi-condition with sort          |
| Data Aggregation      | 4,804 ns       | 2,640 B      | 32 allocs    | Statistical operations             |
| Data Transformation   | 478.7 ns       | 368 B        | 5 allocs     | Field mapping & conversion         |
| Data Validation       | 216.6 ns       | 360 B        | 4 allocs     | Rule-based validation              |
| Stream Processing     | 3,250 ns       | 0 B          | 0 allocs     | Zero-allocation streaming          |
| Large Data Query      | 1.28 ms        | 80 B         | 2 allocs     | 100 records processing             |
| JSON Diff             | 18.2 µs        | 2,787 B      | 197 allocs   | Change detection                   |
| Empty String Handling | 2,777 ns       | 1,664 B      | 27 allocs    | Safe empty string processing       |

### Serialization Performance

| Operation     | Time     | Memory | Allocations | Note                           |
|---------------|----------|--------|-------------|--------------------------------|
| Marshal       | 652.1 ns | 424 B  | 9 allocs    | Standard serialization         |
| FastMarshal   | 226.7 ns | 136 B  | 2 allocs    | High-performance serialization |
| StructMarshal | 267.1 ns | 136 B  | 2 allocs    | Direct struct serialization    |

### Default Value Functions

| Function | Time     | Memory | Allocations |
|----------|----------|--------|-------------|
| StringOr | 23.56 ns | 0 B    | 0 allocs    |
| IntOr    | 28.34 ns | 0 B    | 0 allocs    |
| FloatOr  | 40.89 ns | 0 B    | 0 allocs    |

*Benchmark results on Apple M4 Pro, Go 1.24.6*