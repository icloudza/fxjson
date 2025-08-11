// Package fxjson is a zero-allocation, high-performance JSON parser for Go,
// optimized for fast field/path access and minimal memory overhead.
//
// # Features
//
//   - 0 allocation on core APIs (Get, GetByPath, Len, Index, Exists, etc.).
//   - Unified value receiver for clarity and concurrency safety.
//   - Direct path scanning without building intermediate trees.
//   - Specialized number parsing (Int/Uint/Float/Bool) without strconv.
//   - O(1) array index access via pointer+range cache.
//
// # Example
//
//	b := []byte(`{"data":{"user":{"name":"Alice","age":30,"scores":[99,88,77]}}}`)
//
//	n := fxjson.FromBytes(b)
//	name := n.Get("data").Get("user").Get("name").String() // "Alice"
//	age, _ := n.GetByPath("data.user.age").Int()           // 30
//	s1 := n.GetByPath("data.user.scores").Index(1).NumStr() // "88"
//	ln := n.GetByPath("data.user.scores").Len()            // 3
//	keys := n.Get("data").Get("user").Keys()
//	for _, k := range keys {
//		fmt.Println(string(k))
//	}
//
// # Performance
//
// Benchmarks on Apple M4 Pro (Go 1.24):
//
//	BenchmarkGet_fxjson-12       22.44 ns/op    0 B/op   0 allocs/op
//	BenchmarkGetByPath_fxjson-12 92.77 ns/op    0 B/op   0 allocs/op
//	BenchmarkLen_fxjson-12       18.57 ns/op    0 B/op   0 allocs/op
//
// Full benchmark results in README.
//
// # Notes
//
//   - Assumes valid JSON input (no heavy fault tolerance).
//   - Global array index cache is lock-free and suitable for reused []byte data.
//   - API mirrors gjson for ease of migration but with lower GC noise.
//
// For detailed docs, benchmarks, and examples, see:
//
//	https://github.com/icloudza/fxjson
package fxjson
