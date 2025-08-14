[![Go Version](https://img.shields.io/badge/go-%3E%3D1.18-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Performance](https://img.shields.io/badge/performance-fast-orange.svg)](#performance-comparison)

[📄 中文文档 / Chinese Documentation](README_ZH.md)

FxJSON is a Go JSON parsing library focused on performance, providing efficient JSON traversal and access capabilities. It offers improved performance compared to the standard library while maintaining memory safety and ease of use.

## 🚀 Core Features

- **🔥 Good Performance**: Optimized traversal operations with significant speed improvements
- **⚡ Memory Efficient**: Core operations minimize memory allocations
- **🛡️ Memory Safety**: Proper boundary checking and safety mechanisms
- **🎯 Easy to Use**: Chainable calls with intuitive API design
- **🔧 Feature Complete**: Supports all JSON data types and complex nested structures
- **🌐 Unicode Support**: Handles Chinese, emoji, and other Unicode characters well
- **🧩 Nested JSON Expansion**: Automatic recognition and expansion of nested JSON strings
- **🔢 Number Precision**: Maintains original JSON number formatting with `FloatString()`

## 📊 Performance Comparison

| Operation            | FxJSON   | Standard Library | Performance Gain | Memory Advantage               |
|----------------------|----------|------------------|------------------|--------------------------------|
| ForEach Traversal    | 104.7 ns | 2115 ns          | **20.2x**        | Zero allocations vs 57 allocs  |
| Array Traversal      | 30.27 ns | 2044 ns          | **67.5x**        | Zero allocations vs 57 allocs  |
| Deep Traversal       | 1363 ns  | 2787 ns          | **2.0x**         | 29 allocs vs 83 allocs         |
| Complex Traversal    | 1269 ns  | 3280 ns          | **2.6x**         | Zero allocations vs 104 allocs |
| Large Data Traversal | 11302 ns | 16670 ns         | **1.5x**         | 181 allocs vs 559 allocs       |

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

### Object Traversal (Zero-allocation, 20x performance boost)

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

### Array Traversal (Zero-allocation, 67x performance boost)

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

### Deep Traversal (2x performance boost)

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

## ⚙️ Decode to Struct

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

## ⚠️ Notes

1. **Input Validation**: Assumes valid JSON input, focuses on performance over error handling
2. **Memory Safety**: All string operations include boundary checking
3. **Unicode Support**: Perfect support for Chinese, emoji, and other Unicode characters
4. **Concurrency Safety**: Node read operations are concurrency-safe
5. **Go Version**: Requires Go 1.18 or higher

## 🤝 Contributing

Issues and Pull Requests are welcome!

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details

---

**FxJSON - Make JSON parsing fly!** 🚀