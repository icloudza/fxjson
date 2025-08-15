[![Go Doc](https://img.shields.io/badge/api-reference-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/icloudza/fxjson?utm_source=godoc)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.18-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Performance](https://img.shields.io/badge/performance-fast-orange.svg)](#performance-comparison)

[📄 中文文档 / Chinese Documentation](README_ZH.md)

FxJSON is a Go JSON parsing library focused on performance, providing efficient JSON traversal and access capabilities. It offers improved performance compared to the standard library while maintaining memory safety and ease of use.

## 🚀 Core Features

- **🔥 High Performance**: Optimized traversal operations with significant speed improvements
- **⚡ Memory Efficient**: Core operations minimize memory allocations
- **🛡️ Memory Safety**: Proper boundary checking and safety mechanisms
- **🎯 Easy to Use**: Chainable calls with intuitive API design
- **🔧 Feature Complete**: Supports all JSON data types and complex nested structures
- **🌐 Unicode Support**: Handles Chinese, emoji, and other Unicode characters well
- **🧩 Nested JSON Expansion**: Automatic recognition and expansion of nested JSON strings
- **🔢 Number Precision**: Maintains original JSON number formatting with `FloatString()`
- **🔍 Advanced Querying**: SQL-style conditional queries and filtering
- **📊 Data Aggregation**: Built-in statistical and aggregation functions
- **🎨 Data Transformation**: Flexible field mapping and type conversion
- **✅ Data Validation**: Comprehensive validation rules and sanitization
- **💾 Smart Caching**: High-performance caching with LRU eviction
- **🔧 Debug Tools**: Enhanced debugging and analysis features

## 📊 Performance Comparison

### Core Operations
| Operation            | FxJSON   | Standard Library | Performance Gain | Memory Advantage               |
|----------------------|----------|------------------|------------------|--------------------------------|
| ForEach Traversal    | 104.7 ns | 2115 ns          | **20.2x**        | Zero allocations vs 57 allocs  |
| Array Traversal      | 30.27 ns | 2044 ns          | **67.5x**        | Zero allocations vs 57 allocs  |
| Deep Traversal       | 1363 ns  | 2787 ns          | **2.0x**         | 29 allocs vs 83 allocs         |
| Complex Traversal    | 1269 ns  | 3280 ns          | **2.6x**         | Zero allocations vs 104 allocs |
| Large Data Traversal | 11302 ns | 16670 ns         | **1.5x**         | 181 allocs vs 559 allocs       |

### Advanced Features Performance
| Feature              | Operation Time | Memory Usage | Allocations | Note                    |
|----------------------|----------------|--------------|-------------|-------------------------|
| Basic Parsing        | 5,542 ns       | 6,360 B      | 50 allocs   | Standard JSON parsing   |
| **Cached Parsing**   | **1,396 ns**   | **80 B**     | **3 allocs**| **4x faster, 98% less memory** |
| Data Transformation  | 435 ns         | 368 B        | 5 allocs    | Field mapping & conversion |
| Data Validation      | 208 ns         | 360 B        | 4 allocs    | Rule-based validation   |
| Simple Query         | 2,784 ns       | 640 B        | 14 allocs   | Conditional filtering   |
| Complex Query        | 4,831 ns       | 1,720 B      | 52 allocs   | Multi-condition with sort |
| Data Aggregation     | 4,213 ns       | 2,640 B      | 32 allocs   | Statistical operations  |
| Large Data Query     | 1.27 ms        | 82 B         | 2 allocs    | 100 records processing |
| Stream Processing    | 2,821 ns       | 0 B          | 0 allocs    | Zero-allocation streaming |
| JSON Diff            | 17,200 ns      | 2,710 B      | 197 allocs  | Change detection        |
| Empty String Handling| 3,007 ns       | 1,664 B      | 27 allocs   | Safe empty string processing |

# FxJSON ![Flame](flame.png) - High-Performance JSON Parser

## 📦 Installation

```bash
go get github.com/icloudza/fxjson
```

## 🎯 Quick Start

### Basic Usage

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
        "active": true,
        "score": 95.5,
        "tags": ["developer", "golang"],
        "profile": {
            "city": "Beijing",
            "hobby": "coding"
        }
    }`)

    // Create node
    node := fxjson.FromBytes(jsonData)

    // Basic access
    name, _ := node.Get("name").String()
    age, _ := node.Get("age").Int()
    active, _ := node.Get("active").Bool()
    score, _ := node.Get("score").Float()

    fmt.Printf("Name: %s, Age: %d, Active: %v, Score: %.1f\n", 
               name, age, active, score)
    
    // Nested access
    city, _ := node.Get("profile").Get("city").String()
    fmt.Printf("City: %s\n", city)
    
    // Path access
    hobby, _ := node.GetPath("profile.hobby").String()
    fmt.Printf("Hobby: %s\n", hobby)
}
```

**Output:**
```
Name: Alice, Age: 30, Active: true, Score: 95.5
City: Beijing
Hobby: coding
```

### Array Operations

```go
jsonData := []byte(`{
    "users": [
        {"name": "Alice", "age": 30},
        {"name": "Bob", "age": 25},
        {"name": "Charlie", "age": 35}
    ]
}`)

node := fxjson.FromBytes(jsonData)
users := node.Get("users")

// Array length
fmt.Printf("User count: %d\n", users.Len())

// Index access
firstUser := users.Index(0)
name, _ := firstUser.Get("name").String()
age, _ := firstUser.Get("age").Int()
fmt.Printf("First user: %s (%d years old)\n", name, age)

// Path access to array elements
secondName, _ := node.GetPath("users[1].name").String()
fmt.Printf("Second user: %s\n", secondName)
```

**Output:**
```
User count: 3
First user: Alice (30 years old)
Second user: Bob
```

## 🔄 High-Performance Traversal

### Object Traversal

```go
profile := []byte(`{
    "name": "Developer",
    "skills": ["Go", "Python", "JavaScript"],
    "experience": 5,
    "remote": true
}`)

node := fxjson.FromBytes(profile)

// Zero-allocation high-performance traversal
node.ForEach(func(key string, value fxjson.Node) bool {
    switch value.Kind() {
    case fxjson.TypeString:
        str, _ := value.String()
        fmt.Printf("%s: %s\n", key, str)
    case fxjson.TypeNumber:
        num, _ := value.Int()
        fmt.Printf("%s: %d\n", key, num)
    case fxjson.TypeBool:
        b, _ := value.Bool()
        fmt.Printf("%s: %v\n", key, b)
    case fxjson.TypeArray:
        fmt.Printf("%s: [array, length=%d]\n", key, value.Len())
    }
    return true // continue traversal
})
```

**Output:**
```
name: Developer
skills: [array, length=3]
experience: 5
remote: true
```

### Array Traversal

```go
scores := []byte(`[95, 87, 92, 88, 96]`)
node := fxjson.FromBytes(scores)

var total int64
var count int

// Ultra-fast array traversal (67x performance boost)
node.ArrayForEach(func(index int, value fxjson.Node) bool {
    if score, err := value.Int(); err == nil {
        total += score
        count++
        fmt.Printf("Score %d: %d\n", index+1, score)
    }
    return true
})

fmt.Printf("Average score: %.1f\n", float64(total)/float64(count))
```

**Output:**
```
Score 1: 95
Score 2: 87
Score 3: 92
Score 4: 88
Score 5: 96
Average score: 91.6
```

### Deep Traversal

```go
complexData := []byte(`{
    "company": {
        "name": "Tech Company",
        "departments": [
            {
                "name": "R&D",
                "employees": [
                    {"name": "John", "position": "Engineer"},
                    {"name": "Jane", "position": "Architect"}
                ]
            }
        ]
    }
}`)

node := fxjson.FromBytes(complexData)

// Depth-first traversal of entire JSON tree
node.Walk(func(path string, node fxjson.Node) bool {
    if node.IsString() {
        value, _ := node.String()
        fmt.Printf("Path: %s = %s\n", path, value)
    }
    return true // continue traversing child nodes
})
```

**Output:**
```
Path: company.name = Tech Company
Path: company.departments[0].name = R&D
Path: company.departments[0].employees[0].name = John
Path: company.departments[0].employees[0].position = Engineer
Path: company.departments[0].employees[1].name = Jane
Path: company.departments[0].employees[1].position = Architect
```

## 🛠️ Advanced Features

### Type Checking and Conversion

```go
data := []byte(`{
    "user_id": 12345,
    "username": "developer",
    "is_admin": false,
    "metadata": null,
    "scores": [100, 95, 88]
}`)

node := fxjson.FromBytes(data)

// Type checking
fmt.Printf("user_id is number: %v\n", node.Get("user_id").IsNumber())
fmt.Printf("username is string: %v\n", node.Get("username").IsString())
fmt.Printf("is_admin is bool: %v\n", node.Get("is_admin").IsBool())
fmt.Printf("metadata is null: %v\n", node.Get("metadata").IsNull())
fmt.Printf("scores is array: %v\n", node.Get("scores").IsArray())

// Safe type conversion
if userID, err := node.Get("user_id").Int(); err == nil {
    fmt.Printf("User ID: %d\n", userID)
}

// Get raw JSON
if rawScores := node.Get("scores").Raw(); len(rawScores) > 0 {
    fmt.Printf("Raw scores JSON: %s\n", rawScores)
}
```

**Output:**
```
user_id is number: true
username is string: true
is_admin is bool: true
metadata is null: true
scores is array: true
User ID: 12345
Raw scores JSON: [100, 95, 88]
```

### Number Precision Handling

FxJSON provides special handling for floating-point number precision to maintain the original JSON formatting:

```go
data := []byte(`{
    "price": 1.1,
    "rating": 4.50,
    "score": 95.0,
    "percentage": 12.34
}`)

node := fxjson.FromBytes(data)

// Maintain original JSON number format
price := node.Get("price")
if priceStr, err := price.FloatString(); err == nil {
    fmt.Printf("Price: %s\n", priceStr) // Output: 1.1 (preserves original format)
}

rating := node.Get("rating")
if ratingStr, err := rating.FloatString(); err == nil {
    fmt.Printf("Rating: %s\n", ratingStr) // Output: 4.50 (preserves trailing zero)
}

// Compare with other methods
if floatVal, err := price.Float(); err == nil {
    fmt.Printf("Price as float: %v\n", floatVal)     // Output: 1.1
    fmt.Printf("Price formatted: %g\n", floatVal)    // Output: 1.1
}

// Get original number string
if numStr, err := price.NumStr(); err == nil {
    fmt.Printf("Price NumStr: %s\n", numStr)         // Output: 1.1
}
```

**Output:**
```
Price: 1.1
Rating: 4.50
Price as float: 1.1
Price formatted: 1.1
Price NumStr: 1.1
```

**Methods for number handling:**
- `FloatString()` - Returns original JSON number format (recommended for display)
- `NumStr()` - Returns raw number string from JSON
- `Float()` - Returns `float64` value for calculations
- `Int()` - Returns `int64` value for integers

### Conditional Search and Filtering

```go
students := []byte(`{
    "class": "Advanced Class",
    "students": [
        {"name": "Alice", "grade": 95, "subject": "Math"},
        {"name": "Bob", "grade": 87, "subject": "English"},
        {"name": "Charlie", "grade": 92, "subject": "Math"},
        {"name": "Diana", "grade": 78, "subject": "English"}
    ]
}`)

node := fxjson.FromBytes(students)
studentsArray := node.Get("students")

// Find first Math student
_, student, found := studentsArray.FindInArray(func(index int, value fxjson.Node) bool {
    subject, _ := value.Get("subject").String()
    return subject == "Math"
})

if found {
    name, _ := student.Get("name").String()
    grade, _ := student.Get("grade").Int()
    fmt.Printf("First Math student: %s (grade: %d)\n", name, grade)
}

// Filter all high-score students (>90)
highScoreStudents := studentsArray.FilterArray(func(index int, value fxjson.Node) bool {
    grade, _ := value.Get("grade").Int()
    return grade > 90
})

fmt.Printf("High-score students count: %d\n", len(highScoreStudents))
for i, student := range highScoreStudents {
    name, _ := student.Get("name").String()
    grade, _ := student.Get("grade").Int()
    fmt.Printf("High-score student %d: %s (%d points)\n", i+1, name, grade)
}
```

**Output:**
```
First Math student: Alice (grade: 95)
High-score students count: 2
High-score student 1: Alice (95 points)
High-score student 2: Charlie (92 points)
```

### Statistics and Analysis

```go
data := []byte(`{
    "sales": [
        {"amount": 1500, "region": "North"},
        {"amount": 2300, "region": "South"},
        {"amount": 1800, "region": "North"},
        {"amount": 2100, "region": "South"}
    ]
}`)

node := fxjson.FromBytes(data)
salesArray := node.Get("sales")

// Count North region sales
northCount := salesArray.CountIf(func(index int, value fxjson.Node) bool {
    region, _ := value.Get("region").String()
    return region == "North"
})

// Check if all sales are above 1000
allAbove1000 := salesArray.AllMatch(func(index int, value fxjson.Node) bool {
    amount, _ := value.Get("amount").Int()
    return amount > 1000
})

// Check if any sales exceed 2000
hasHighSales := salesArray.AnyMatch(func(index int, value fxjson.Node) bool {
    amount, _ := value.Get("amount").Int()
    return amount > 2000
})

fmt.Printf("North region records: %d\n", northCount)
fmt.Printf("All above 1000: %v\n", allAbove1000)
fmt.Printf("Has sales > 2000: %v\n", hasHighSales)
```

**Output:**
```
North region records: 2
All above 1000: true
Has sales > 2000: true
```

## 🌟 Complex Application Scenarios

### Nested JSON String Processing

```go
// Data containing nested JSON strings
complexJSON := []byte(`{
    "user_info": "{\"name\":\"John\",\"age\":30,\"skills\":[\"Go\",\"Python\"]}",
    "config": "{\"theme\":\"dark\",\"language\":\"en-US\"}",
    "regular_field": "regular string"
}`)

node := fxjson.FromBytes(complexJSON)

// FxJSON automatically recognizes and expands nested JSON strings
userInfo := node.Get("user_info")
if userInfo.IsObject() { // Nested JSON is automatically expanded to object
    name, _ := userInfo.Get("name").String()
    age, _ := userInfo.Get("age").Int()
    fmt.Printf("User: %s, Age: %d\n", name, age)
    
    // Traverse skills array
    fmt.Print("Skills: ")
    userInfo.Get("skills").ArrayForEach(func(index int, skill fxjson.Node) bool {
        skillName, _ := skill.String()
        fmt.Printf("%s ", skillName)
        return true
    })
    fmt.Println()
}

// Config is also automatically expanded
config := node.Get("config")
if config.IsObject() {
    theme, _ := config.Get("theme").String()
    language, _ := config.Get("language").String()
    fmt.Printf("Theme: %s, Language: %s\n", theme, language)
}
```

**Output:**
```
User: John, Age: 30
Skills: Go Python 
Theme: dark, Language: en-US
```

### Configuration File Parsing

```go
configJSON := []byte(`{
    "database": {
        "host": "localhost",
        "port": 5432,
        "name": "myapp",
        "ssl": true,
        "pool": {
            "min": 5,
            "max": 100
        }
    },
    "redis": {
        "host": "127.0.0.1",
        "port": 6379,
        "db": 0
    },
    "features": ["auth", "logging", "metrics"]
}`)

config := fxjson.FromBytes(configJSON)

// Database configuration
dbHost, _ := config.GetPath("database.host").String()
dbPort, _ := config.GetPath("database.port").Int()
sslEnabled, _ := config.GetPath("database.ssl").Bool()
maxPool, _ := config.GetPath("database.pool.max").Int()

fmt.Printf("Database: %s:%d (SSL: %v, Max Pool: %d)\n", 
           dbHost, dbPort, sslEnabled, maxPool)

// Redis configuration
redisHost, _ := config.GetPath("redis.host").String()
redisPort, _ := config.GetPath("redis.port").Int()
fmt.Printf("Redis: %s:%d\n", redisHost, redisPort)

// Feature list
features := config.Get("features")
fmt.Printf("Enabled features (%d items): ", features.Len())
features.ArrayForEach(func(index int, feature fxjson.Node) bool {
    name, _ := feature.String()
    fmt.Printf("%s ", name)
    return true
})
fmt.Println()
```

**Output:**
```
Database: localhost:5432 (SSL: true, Max Pool: 100)
Redis: 127.0.0.1:6379
Enabled features (3 items): auth logging metrics 
```

## ⚙️ High-Performance Decode to Struct

FxJSON provides multiple optimized decoding methods for different performance requirements:

### Standard Decode (Node-based)

```go
type User struct {
    Name  string   `json:"name"`
    Age   int      `json:"age"`
    Tags  []string `json:"tags"`
    Email string   `json:"email"`
}

jsonData := []byte(`{
    "name": "Developer",
    "age": 28,
    "tags": ["golang", "json", "performance"],
    "email": "dev@example.com"
}`)

node := fxjson.FromBytes(jsonData)

var user User
if err := node.Decode(&user); err != nil {
    fmt.Printf("Decode error: %v\n", err)
} else {
    fmt.Printf("Decode result:\n")
    fmt.Printf("  Name: %s\n", user.Name)
    fmt.Printf("  Age: %d\n", user.Age)
    fmt.Printf("  Email: %s\n", user.Email)
    fmt.Printf("  Tags: %v\n", user.Tags)
}
```

**Output:**
```
Decode result:
  Name: Developer
  Age: 28
  Email: dev@example.com
  Tags: [golang json performance]
```

### Direct Decode (Optimized)

For better performance, you can decode directly from bytes without creating a Node:

```go
// DecodeStruct - Direct decoding from bytes (faster)
var user1 User
if err := fxjson.DecodeStruct(jsonData, &user1); err != nil {
    fmt.Printf("DecodeStruct error: %v\n", err)
} else {
    fmt.Printf("DecodeStruct result: %+v\n", user1)
}

// DecodeStructFast - Ultra-fast decoding (fastest)
var user2 User
if err := fxjson.DecodeStructFast(jsonData, &user2); err != nil {
    fmt.Printf("DecodeStructFast error: %v\n", err)
} else {
    fmt.Printf("DecodeStructFast result: %+v\n", user2)
}
```

**Output:**
```
DecodeStruct result: {Name:Developer Age:28 Tags:[golang json performance] Email:dev@example.com}
DecodeStructFast result: {Name:Developer Age:28 Tags:[golang json performance] Email:dev@example.com}
```

### Performance Comparison

| Method | Speed | Use Case |
|--------|-------|----------|
| `node.Decode()` | Fast | When you need Node functionality |
| `DecodeStruct()` | Faster | Direct struct decoding |
| `DecodeStructFast()` | Fastest | Performance-critical scenarios |

### Complex Struct Decoding

```go
type ComplexUser struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Profile  struct {
        Avatar string   `json:"avatar"`
        Bio    string   `json:"bio"`
        Skills []string `json:"skills"`
    } `json:"profile"`
    Metadata map[string]interface{} `json:"metadata"`
}

complexJSON := []byte(`{
    "id": 12345,
    "name": "Advanced Developer",
    "profile": {
        "avatar": "https://example.com/avatar.jpg",
        "bio": "Full-stack developer with 10+ years experience",
        "skills": ["Go", "Python", "JavaScript", "Docker", "Kubernetes"]
    },
    "metadata": {
        "last_login": "2024-01-15T10:30:00Z",
        "preferences": {
            "theme": "dark",
            "language": "en-US"
        }
    }
}`)

var complexUser ComplexUser
if err := fxjson.DecodeStructFast(complexJSON, &complexUser); err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

fmt.Printf("User ID: %d\n", complexUser.ID)
fmt.Printf("Name: %s\n", complexUser.Name)
fmt.Printf("Bio: %s\n", complexUser.Profile.Bio)
fmt.Printf("Skills: %v\n", complexUser.Profile.Skills)
fmt.Printf("Metadata: %+v\n", complexUser.Metadata)
```

**Output:**
```
User ID: 12345
Name: Advanced Developer
Bio: Full-stack developer with 10+ years experience
Skills: [Go Python JavaScript Docker Kubernetes]
Metadata: map[last_login:2024-01-15T10:30:00Z preferences:map[language:en-US theme:dark]]
```

## 🚨 Error Handling

```go
jsonData := []byte(`{
    "number": "not_a_number",
    "missing": null,
    "empty_string": "",
    "valid_number": 42
}`)

node := fxjson.FromBytes(jsonData)

// Type conversion error handling
if num, err := node.Get("number").Int(); err != nil {
    fmt.Printf("Number conversion failed: %v\n", err)
}

// Successful type conversion
if num, err := node.Get("valid_number").Int(); err == nil {
    fmt.Printf("Valid number: %d\n", num)
}

// Check if field exists
if node.HasKey("missing_field") {
    fmt.Println("missing_field exists")
} else {
    fmt.Println("missing_field does not exist")
}

if node.HasKey("valid_number") {
    fmt.Println("valid_number exists")
}

// Use default value
defaultNode := fxjson.FromBytes([]byte(`"default_value"`))
value := node.GetKeyValue("missing_field", defaultNode)
defaultStr, _ := value.String()
fmt.Printf("Using default value: %s\n", defaultStr)

// Handle empty string
emptyStr, err := node.Get("empty_string").String()
if err == nil {
    fmt.Printf("Empty string length: %d\n", len(emptyStr))
}
```

**Output:**
```
Number conversion failed: node is not a number type (got type="string")
Valid number: 42
missing_field does not exist
valid_number exists
Using default value: default_value
Empty string length: 0
```

## 🎨 Convenience Methods

```go
data := []byte(`{
    "company": {
        "name": "Tech Company",
        "founded": 2020,
        "employees": [
            {"name": "John", "department": "R&D", "salary": 15000},
            {"name": "Jane", "department": "Marketing", "salary": 12000},
            {"name": "Bob", "department": "R&D", "salary": 18000}
        ]
    }
}`)

node := fxjson.FromBytes(data)

// Convert to Map
fmt.Println("=== Company Info (ToMap) ===")
companyMap := node.Get("company").ToMap()
for key, value := range companyMap {
    if key == "employees" {
        fmt.Printf("%s: [array, length=%d]\n", key, value.Len())
    } else {
        fmt.Printf("%s: %s\n", key, string(value.Raw()))
    }
}

// Convert to Slice
fmt.Println("\n=== Employee List (ToSlice) ===")
employees := node.GetPath("company.employees").ToSlice()
fmt.Printf("Total employees: %d\n", len(employees))
for i, employee := range employees {
    name, _ := employee.Get("name").String()
    dept, _ := employee.Get("department").String()
    salary, _ := employee.Get("salary").Int()
    fmt.Printf("Employee %d: %s - %s dept (salary: %d)\n", i+1, name, dept, salary)
}

// Get all keys
fmt.Println("\n=== Company Fields (GetAllKeys) ===")
keys := node.Get("company").GetAllKeys()
fmt.Printf("Company fields: %v\n", keys)

// Get all employee nodes
fmt.Println("\n=== Employee Nodes (GetAllValues) ===")
employeeNodes := node.GetPath("company.employees").GetAllValues()
fmt.Printf("Employee node count: %d\n", len(employeeNodes))
for i, empNode := range employeeNodes {
    name, _ := empNode.Get("name").String()
    fmt.Printf("Node %d: %s's info\n", i+1, name)
}
```

**Output:**
```
=== Company Info (ToMap) ===
name: "Tech Company"
founded: 2020
employees: [array, length=3]

=== Employee List (ToSlice) ===
Total employees: 3
Employee 1: John - R&D dept (salary: 15000)
Employee 2: Jane - Marketing dept (salary: 12000)
Employee 3: Bob - R&D dept (salary: 18000)

=== Company Fields (GetAllKeys) ===
Company fields: [name founded employees]

=== Employee Nodes (GetAllValues) ===
Employee node count: 3
Node 1: John's info
Node 2: Jane's info
Node 3: Bob's info
```

## 📝 Performance Tips

1. **Traversal Optimization**: For large datasets, prefer `ForEach`, `ArrayForEach`, and `Walk` methods
2. **Path Access**: Use `GetPath` for one-shot access to deeply nested fields
3. **Memory Management**: Core traversal operations achieve zero allocations, suitable for high-frequency scenarios
4. **Type Checking**: Use `IsXXX()` methods for type checking to avoid unnecessary type conversions
5. **Cache Utilization**: Array indices are automatically cached for better performance on repeated access
6. **Decode Optimization**: 
   - Use `node.Decode()` when you need Node functionality
   - Use `DecodeStruct()` for direct struct decoding (faster)
   - Use `DecodeStructFast()` for performance-critical scenarios (fastest)
   - Choose the right method based on your performance requirements

## ⚠️ Notes

1. **Input Validation**: Assumes valid JSON input, focuses on performance over error handling
2. **Memory Safety**: All string operations include boundary checking
3. **Unicode Support**: Perfect support for Chinese, emoji, and other Unicode characters
4. **Concurrency Safety**: Node read operations are concurrency-safe
5. **Go Version**: Requires Go 1.18 or higher

## 📚 Complete API Reference

### Core Methods

#### Node Creation
- `FromBytes(data []byte) Node` - Create node from JSON bytes with automatic nested JSON expansion

#### Basic Access
- `Get(key string) Node` - Get object field by key
- `GetPath(path string) Node` - Get value by path (e.g., "user.profile.name")
- `Index(i int) Node` - Get array element by index

#### Type Checking
- `Exists() bool` - Check if node exists
- `IsObject() bool` - Check if node is JSON object
- `IsArray() bool` - Check if node is JSON array
- `IsString() bool` - Check if node is JSON string
- `IsNumber() bool` - Check if node is JSON number
- `IsBool() bool` - Check if node is JSON boolean
- `IsNull() bool` - Check if node is JSON null
- `IsScalar() bool` - Check if node is scalar type (string, number, bool, null)
- `IsContainer() bool` - Check if node is container type (object, array)
- `Kind() NodeType` - Get node type as enum
- `Type() byte` - Get internal type byte

#### Value Extraction
- `String() (string, error)` - Get string value
- `Int() (int64, error)` - Get integer value
- `Uint() (uint64, error)` - Get unsigned integer value
- `Float() (float64, error)` - Get floating-point value
- `Bool() (bool, error)` - Get boolean value
- `NumStr() (string, error)` - Get raw number string from JSON
- `FloatString() (string, error)` - Get number string preserving original JSON format
- `Raw() []byte` - Get raw JSON bytes for this node
- `RawString() (string, error)` - Get raw JSON as string
- `Json() (string, error)` - Get JSON representation (objects/arrays only)

#### Size and Keys
- `Len() int` - Get length (array elements, object fields, string characters)
- `Keys() [][]byte` - Get object keys as byte slices
- `GetAllKeys() []string` - Get object keys as strings
- `GetAllValues() []Node` - Get array elements as nodes
- `ToMap() map[string]Node` - Convert object to map
- `ToSlice() []Node` - Convert array to slice

#### High-Performance Traversal
- `ForEach(fn ForEachFunc) bool` - Traverse object with zero allocations (20x faster)
- `ArrayForEach(fn ArrayForEachFunc) bool` - Traverse array with zero allocations (67x faster)
- `Walk(fn WalkFunc) bool` - Deep traversal of entire JSON tree (2x faster)

#### Search and Filter
- `FindInObject(predicate func(key string, value Node) bool) (string, Node, bool)` - Find first matching object field
- `FindInArray(predicate func(index int, value Node) bool) (int, Node, bool)` - Find first matching array element
- `FilterArray(predicate func(index int, value Node) bool) []Node` - Filter array elements
- `FindByPath(path string) Node` - Alias for GetPath

#### Conditional Operations
- `HasKey(key string) bool` - Check if object has key
- `GetKeyValue(key string, defaultValue Node) Node` - Get value with default fallback
- `CountIf(predicate func(index int, value Node) bool) int` - Count matching array elements
- `AllMatch(predicate func(index int, value Node) bool) bool` - Check if all array elements match
- `AnyMatch(predicate func(index int, value Node) bool) bool` - Check if any array element matches

#### Decoding
- `Decode(v any) error` - Decode JSON into Go struct/type (optimized)
- `DecodeStruct(data []byte, v any) error` - Direct struct decoding from bytes
- `DecodeStructFast(data []byte, v any) error` - Ultra-fast struct decoding

### Callback Function Types

```go
// Object traversal callback
type ForEachFunc func(key string, value Node) bool

// Array traversal callback  
type ArrayForEachFunc func(index int, value Node) bool

// Deep traversal callback
type WalkFunc func(path string, node Node) bool
```

### Node Types

```go
const (
    TypeInvalid NodeType = 0
    TypeObject  NodeType = 'o'
    TypeArray   NodeType = 'a' 
    TypeString  NodeType = 's'
    TypeNumber  NodeType = 'n'
    TypeBool    NodeType = 'b'
    TypeNull    NodeType = 'l'
)
```

## 🤝 Contributing

Issues and Pull Requests are welcome!

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details

---

## 🔍 Advanced Features

### SQL-Style Querying

```go
notesData := []byte(`{
    "notes": [
        {"id": "1", "title": "Go Tutorial", "views": 1250, "category": "tech"},
        {"id": "2", "title": "Cooking Tips", "views": 890, "category": "food"},
        {"id": "3", "title": "Travel Guide", "views": 2100, "category": "travel"}
    ]
}`)

node := fxjson.FromBytes(notesData)
notesList := node.Get("notes")

// Complex query with multiple conditions
results, err := notesList.Query().
    Where("views", ">", 1000).
    Where("category", "!=", "food").
    SortBy("views", "desc").
    Limit(10).
    ToSlice()

if err == nil {
    fmt.Printf("Found %d high-view notes\n", len(results))
    for _, note := range results {
        title, _ := note.Get("title").String()
        views, _ := note.Get("views").Int()
        fmt.Printf("- %s (%d views)\n", title, views)
    }
}
```

**Output:**
```
Found 2 high-view notes
- Travel Guide (2100 views)
- Go Tutorial (1250 views)
```

### Data Aggregation & Statistics

```go
// Group by category and calculate statistics
stats, err := notesList.Aggregate().
    GroupBy("category").
    Count("total_notes").
    Sum("views", "total_views").
    Avg("views", "avg_views").
    Max("views", "max_views").
    Execute(notesList)

if err == nil {
    fmt.Println("Statistics by Category:")
    for category, data := range stats {
        statsMap := data.(map[string]interface{})
        fmt.Printf("📁 %s: %d notes, %.0f total views, %.1f avg views\n",
            category, int(statsMap["total_notes"].(float64)),
            statsMap["total_views"], statsMap["avg_views"])
    }
}
```

**Output:**
```
Statistics by Category:
📁 tech: 1 notes, 1250 total views, 1250.0 avg views
📁 food: 1 notes, 890 total views, 890.0 avg views
📁 travel: 1 notes, 2100 total views, 2100.0 avg views
```

### Data Transformation & Mapping

```go
// Transform data structure with field mapping
mapper := fxjson.FieldMapper{
    Rules: map[string]string{
        "notes[0].title": "post_title",
        "notes[0].views": "view_count",
        "notes[0].category": "post_category",
    },
    DefaultValues: map[string]interface{}{
        "status": "published",
        "created_by": "system",
    },
    TypeCast: map[string]string{
        "view_count": "int",
    },
}

result, err := node.Transform(mapper)
if err == nil {
    fmt.Println("Transformed data:")
    for key, value := range result {
        fmt.Printf("  %s: %v\n", key, value)
    }
}
```

**Output:**
```
Transformed data:
  post_title: Go Tutorial
  view_count: 1250
  post_category: tech
  status: published
  created_by: system
```

### High-Performance Caching

```go
// Enable caching for better performance
cache := fxjson.NewMemoryCache(100)
fxjson.EnableCaching(cache)

// First parse (cache miss)
start := time.Now()
node1 := fxjson.FromBytesWithCache(notesData, 5*time.Minute)
firstTime := time.Since(start)

// Second parse (cache hit)
start = time.Now()
node2 := fxjson.FromBytesWithCache(notesData, 5*time.Minute)
secondTime := time.Since(start)

stats := cache.Stats()
fmt.Printf("First parse: %v\n", firstTime)
fmt.Printf("Cached parse: %v (%.1fx faster)\n", 
    secondTime, float64(firstTime)/float64(secondTime))
fmt.Printf("Cache hit rate: %.1f%%\n", stats.HitRate*100)
```

**Output:**
```
First parse: 45.2µs
Cached parse: 4.8µs (9.4x faster)
Cache hit rate: 50.0%
```

### Data Validation

```go
// Define validation rules
validator := &fxjson.DataValidator{
    Rules: map[string]fxjson.ValidationRule{
        "title": {
            Required:  true,
            Type:      "string",
            MinLength: 1,
            MaxLength: 100,
        },
        "views": {
            Required: true,
            Type:     "number",
            Min:      0,
            Max:      1000000,
        },
    },
}

// Validate first note
firstNote := notesList.Index(0)
result, errors := firstNote.Validate(validator)

if len(errors) == 0 {
    fmt.Println("✅ Validation passed")
    fmt.Printf("Validated fields: %d\n", len(result))
} else {
    fmt.Println("❌ Validation failed:")
    for _, err := range errors {
        fmt.Printf("  - %s\n", err)
    }
}
```

### Enhanced Debugging

```go
// Enable debug mode
fxjson.EnableDebugMode()
defer fxjson.DisableDebugMode()

// Parse with debug information
node, debugInfo := fxjson.FromBytesWithDebug(notesData)

fmt.Printf("📊 Debug Information:\n")
fmt.Printf("  Parse time: %v\n", debugInfo.ParseTime)
fmt.Printf("  Memory usage: %d bytes\n", debugInfo.MemoryUsage)
fmt.Printf("  Node count: %d\n", debugInfo.NodeCount)
fmt.Printf("  Max depth: %d\n", debugInfo.MaxDepth)

// Pretty print JSON structure
prettyOutput := node.PrettyPrint()
fmt.Printf("\n📝 Pretty JSON:\n%s\n", prettyOutput)

// Analyze JSON structure
inspection := node.Inspect()
fmt.Printf("\n🔍 Structure Analysis:\n")
fmt.Printf("  Type: %v\n", inspection["type"])
fmt.Printf("  Key count: %v\n", inspection["key_count"])
```

**Output:**
```
📊 Debug Information:
  Parse time: 125.4µs
  Memory usage: 15360 bytes
  Node count: 42
  Max depth: 3

📝 Pretty JSON:
{
  "notes": [
    {
      "id": "1",
      "title": "Go Tutorial",
      "views": 1250,
      "category": "tech"
    },
    ...
  ]
}

🔍 Structure Analysis:
  Type: 111
  Key count: 1
```

### Stream Processing & Batch Operations

```go
// Stream processing for large datasets
processedCount := 0
err := notesList.Stream(func(note fxjson.Node, index int) bool {
    title, _ := note.Get("title").String()
    views, _ := note.Get("views").Int()
    
    fmt.Printf("Processing note %d: %s (%d views)\n", index+1, title, views)
    processedCount++
    
    // Return false to stop early if needed
    return true
})

fmt.Printf("Processed %d notes via streaming\n", processedCount)

// Batch processing with custom batch size
batchProcessor := fxjson.NewBatchProcessor(2, func(nodes []fxjson.Node) error {
    fmt.Printf("Processing batch of %d nodes\n", len(nodes))
    // Process batch...
    return nil
})

notesList.ArrayForEach(func(index int, note fxjson.Node) bool {
    batchProcessor.Add(note)
    return true
})
batchProcessor.Flush()
```

**Output:**
```
Processing note 1: Go Tutorial (1250 views)
Processing note 2: Cooking Tips (890 views)
Processing note 3: Travel Guide (2100 views)
Processed 3 notes via streaming
Processing batch of 2 nodes
Processing batch of 1 nodes
```

## 🎯 Use Cases

### 1. **Configuration Management**
- Complex configuration parsing with validation
- Environment-specific configuration merging
- Real-time configuration updates with caching

### 2. **API Response Processing**
- High-throughput API response parsing
- Data transformation for different API versions
- Response filtering and aggregation

### 3. **Data Analytics**
- Large dataset analysis and aggregation
- Real-time metrics calculation
- Data quality validation and sanitization

### 4. **Content Management**
- Document structure analysis
- Content transformation and migration
- Search and filtering operations

### 5. **Log Processing**
- Structured log parsing and analysis
- Log aggregation and statistics
- Performance monitoring and debugging

**FxJSON - Make JSON parsing fly!** 🚀