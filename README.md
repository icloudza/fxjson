# fxjson

A zero-allocation, high-performance JSON parser for Go, focusing on **fast path access** and **ultra-low memory overhead**.

[📄 中文文档 / Chinese Documentation](README_ZH.md)

> Goal: Maintain API simplicity while minimizing latency and allocations, tailored for high-QPS, low-jitter scenarios.

---

## ✨ Features

* **0 Allocation**: Core APIs such as `Get` / `GetPath` / `Len` / `Index` are zero-allocation.
* **Unified Value Receivers**: Complies with Go best practices to avoid value/pointer receiver ambiguity and escape analysis pitfalls.
* **Path Direct Access**: No intermediate tree building; scans raw byte streams directly and skips irrelevant fields.
* **Specialized Number Parsing**: Hand-written `Int/Uint/Float/Bool` implementations without `strconv`, with extremely short call stacks.
* **O(1) Array Indexing**: Uses `(data pointer + slice range)` global lock-free cache for extremely fast repeated indexing.

---

## 🚀 Installation

```bash
go get github.com/icloudza/fxjson
```

---

## 🔧 Quick Start

```go
package main

import (
	"fmt"
	"github.com/icloudza/fxjson"
)

func main() {
	b := []byte(`{"data":{"user":{"name":"Alice","age":30,"scores":[99,88,77]}}}`)

	n := fxjson.FromBytes(b)
	name := n.Get("data").Get("user").Get("name").String() // "Alice"
	fmt.Println("name:", name)

	age, _ := n.GetPath("data.user.age").Int() // 30
	fmt.Println("age:", age)

	s1, _ := n.GetPath("data.user.scores").Index(1).NumStr() // "88"
	fmt.Println("s1:", s1)

	ln := n.GetPath("data.user.scores").Len() // 3
	fmt.Println("ln:", ln)

	keys := n.Get("data").Get("user").Keys() // [][]byte{"name","age","scores"}
	for i, k := range keys {
		fmt.Printf("user index %d %s \n", i, string(k))
	}
}

//result:
//name: Alice
//age: 30
//s1: 88
//ln: 3
//user index 0 name
//user index 1 age
//user index 2 scores 
```

**Output**

```
name: Alice
age: 30
s1: 88
ln: 3
user index 0 name
user index 1 age
user index 2 scores
```

---

## 📊 Benchmark (Apple M4 Pro)

Command:

```bash
go test -bench . -benchmem -cpuprofile=cpu.out
```

| Benchmark        | fxjson ns/op | fxjson B/op | fxjson allocs/op | gjson ns/op | gjson B/op | gjson allocs/op |
|------------------|--------------|-------------|------------------|-------------|------------|-----------------|
| BenchmarkGet     | 22.44        | 0           | 0                | 45.60       | 8          | 1               |
| BenchmarkGetPath | 92.77        | 0           | 0                | 138.5       | 5          | 1               |
| BenchmarkInt     | 14.84        | 0           | 0                | 9.134       | 0          | 0               |
| BenchmarkFloat   | 6.684        | 0           | 0                | 1.893       | 0          | 0               |
| BenchmarkBool    | 1.788        | 0           | 0                | 1.878       | 0          | 0               |
| BenchmarkString  | 0.9925       | 0           | 0                | 2.004       | 0          | 0               |
| BenchmarkNumStr  | 0.8289       | 0           | 0                | 0.2287      | 0          | 0               |
| BenchmarkLen     | 18.57        | 0           | 0                | 130.6       | 560        | 3               |
| BenchmarkKeys    | 114.7        | 168         | 3                | 213.1       | 944        | 2               |
| BenchmarkIndex   | 14.00        | 0           | 0                | 0.2255      | 0          | 0               |
| BenchmarkExists  | 0.2261       | 0           | 0                | 1.301       | 0          | 0               |
| BenchmarkIsNull  | 0.2247       | 0           | 0                | 0.2266      | 0          | 0               |
| BenchmarkDecode  | 129.8        | 368         | 5                | 108.5       | 0          | 0               |

---

## 🔥 Flamegraph (pprof)

Generated:

```
flame.png
```
![CPU Flamegraph](flame.png)

### Reproduce

```bash
# Run benchmark and generate CPU profile
go test -bench . -benchmem -cpuprofile=cpu.out

# Generate SVG (best readability)
go tool pprof -svg ./fxjson.test cpu.out > flame.svg

# Optionally PNG/JPG:
# 1) Go 1.22+ try -png (if supported)
# go tool pprof -png ./fxjson.test cpu.out > flame.png
# 2) Convert with ImageMagick
# convert flame.svg flame.png

# Web interface (interactive)
go tool pprof -http=:8080 ./fxjson.test cpu.out
```

---

## 🆚 Advantages over gjson

* **Lower GC noise**: Core APIs are zero-allocation; flamegraphs show almost no `mallocgc`/`makeslice`.
* **Shallower stack & focused hotspots**: `findObjectField` / `skipValueFast` are single hotspots, making further optimization easier.
* **Optimized path parsing**: `GetPath` directly scans bytes, avoiding generic branches and handling complex objects efficiently.
* **Custom integer/float parsing**: No `strconv`, short execution path, higher throughput.
* **O(1) Array Indexing**: Repeated access to the same array is significantly faster.

> Suitable for: API gateways, log/metrics ingestion, real-time risk control, trading/matching systems.

---

## 📚 API Overview

```go
n := fxjson.FromBytes(data)

// Field / path access
n.Get("data")
n.GetPath("a.b[3].c")

// Arrays
n.Get("arr").Len()
n.Get("arr").Index(0)

// Primitive types
n.Get("s").String()
n.Get("i").Int()
n.Get("u").Uint()
n.Get("f").Float()
n.Get("b").Bool()

// Other utilities
n.Exists()
n.IsNull()
n.NumStr() // raw numeric string (zero-alloc)
n.Raw()    // raw JSON slice (zero-copy)

// Decode to any (quick & easy)
var v any
_ = n.Decode(&v)
```

---

## ⚠️ Notes

* Input is assumed to be **valid JSON**; the parser is optimized for performance, not heavy error handling.
* Global array index cache uses `sync.Map` keyed by `(data pointer + range)`; suitable for long-lived raw byte reuse (e.g., log batches, bulk processing).
* Unified value receivers: Does not modify `Node` state, safe for chained calls and concurrent reads.

---

## 📦 Compatibility

* Go 1.21+

---

## 📄 License

MIT
