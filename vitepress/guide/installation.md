# 安装配置

本节详细介绍如何在您的 Go 项目中安装和配置 FxJSON。

## 系统要求

- **Go 版本**: 1.18 或更高版本
- **操作系统**: 支持 Linux、macOS、Windows
- **架构**: 支持 amd64、arm64 等主流架构

## 安装方法

### 使用 go get（推荐）

这是最简单的安装方法：

```bash
go get github.com/icloudza/fxjson
```

### 在 go.mod 中添加依赖

您也可以直接在 `go.mod` 文件中添加依赖：

```go
module your-project

go 1.18

require (
    github.com/icloudza/fxjson latest
)
```

然后运行：

```bash
go mod tidy
```

### 指定版本

如果您需要特定版本：

```bash
# 安装特定版本
go get github.com/icloudza/fxjson@v1.0.0

# 或在 go.mod 中指定
require github.com/icloudza/fxjson v1.0.0
```

## 验证安装

创建一个简单的测试文件来验证安装：

```go
// test_install.go
package main

import (
    "fmt"
    "github.com/icloudza/fxjson"
)

func main() {
    // 测试基本功能
    jsonStr := `{"message": "FxJSON 安装成功!", "version": "1.0"}`
    node := fxjson.FromBytes([]byte(jsonStr))
    
    message := node.Get("message").StringOr("安装失败")
    version := node.Get("version").StringOr("unknown")
    
    fmt.Printf("消息: %s\n", message)
    fmt.Printf("版本: %s\n", version)
}
```

运行测试：

```bash
go run test_install.go
```

预期输出：
```
消息: FxJSON 安装成功!
版本: 1.0
```

## 项目配置

### 基本项目结构

```
your-project/
├── go.mod
├── go.sum
├── main.go
├── config/
│   └── config.go
├── models/
│   └── user.go
└── handlers/
    └── api.go
```

### 配置示例

#### config/config.go

```go
package config

import (
    "os"
    "github.com/icloudza/fxjson"
)

type Config struct {
    Server   ServerConfig   `json:"server"`
    Database DatabaseConfig `json:"database"`
    Redis    RedisConfig    `json:"redis"`
}

type ServerConfig struct {
    Port int    `json:"port"`
    Host string `json:"host"`
}

type DatabaseConfig struct {
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Database string `json:"database"`
    Username string `json:"username"`
    Password string `json:"password"`
}

type RedisConfig struct {
    Host string `json:"host"`
    Port int    `json:"port"`
    DB   int    `json:"db"`
}

func LoadConfig(filepath string) (*Config, error) {
    data, err := os.ReadFile(filepath)
    if err != nil {
        return nil, err
    }
    
    node := fxjson.FromBytes(data)
    
    var config Config
    err = node.DecodeStruct(&config)
    if err != nil {
        return nil, err
    }
    
    return &config, nil
}

// 使用安全访问方式
func LoadConfigSafe(filepath string) *Config {
    data, err := os.ReadFile(filepath)
    if err != nil {
        return getDefaultConfig()
    }
    
    node := fxjson.FromBytes(data)
    
    return &Config{
        Server: ServerConfig{
            Host: node.GetPath("server.host").StringOr("localhost"),
            Port: int(node.GetPath("server.port").IntOr(8080)),
        },
        Database: DatabaseConfig{
            Host:     node.GetPath("database.host").StringOr("localhost"),
            Port:     int(node.GetPath("database.port").IntOr(5432)),
            Database: node.GetPath("database.database").StringOr("myapp"),
            Username: node.GetPath("database.username").StringOr("user"),
            Password: node.GetPath("database.password").StringOr(""),
        },
        Redis: RedisConfig{
            Host: node.GetPath("redis.host").StringOr("localhost"),
            Port: int(node.GetPath("redis.port").IntOr(6379)),
            DB:   int(node.GetPath("redis.db").IntOr(0)),
        },
    }
}

func getDefaultConfig() *Config {
    return &Config{
        Server:   ServerConfig{Host: "localhost", Port: 8080},
        Database: DatabaseConfig{Host: "localhost", Port: 5432, Database: "myapp"},
        Redis:    RedisConfig{Host: "localhost", Port: 6379, DB: 0},
    }
}
```

#### config.json

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080
  },
  "database": {
    "host": "localhost",
    "port": 5432,
    "database": "myapp",
    "username": "postgres",
    "password": "password"
  },
  "redis": {
    "host": "localhost",
    "port": 6379,
    "db": 0
  }
}
```

## 性能配置

### 解析选项配置

对于特殊需求，您可以配置解析选项：

```go
package main

import (
    "github.com/icloudza/fxjson"
)

func main() {
    // 自定义解析选项
    opts := fxjson.ParseOptions{
        MaxDepth:      100,         // 限制最大嵌套深度
        MaxStringLen:  1024 * 1024, // 限制字符串最大长度
        MaxObjectKeys: 10000,       // 限制对象最大键数
        MaxArrayItems: 100000,      // 限制数组最大元素数
        StrictMode:    true,        // 启用严格模式
    }
    
    data := []byte(`{"key": "value"}`)
    node := fxjson.FromBytesWithOptions(data, opts)
    
    // 使用节点...
}
```

### 缓存配置

对于需要重复访问的大型 JSON：

```go
func setupCache() {
    // 加载大型配置文件
    configData, _ := os.ReadFile("large_config.json")
    
    // 解析配置文件
    config := fxjson.FromBytes(configData)
    
    // 重复访问时会使用缓存
    for i := 0; i < 1000; i++ {
        _ = config.GetPath("app.features.auth.enabled").BoolOr(false)
    }
}
```

## 最佳实践配置

### 1. 错误处理配置

```go
// 定义自定义错误类型
type ConfigError struct {
    Field   string
    Message string
}

func (e *ConfigError) Error() string {
    return fmt.Sprintf("配置错误 [%s]: %s", e.Field, e.Message)
}

// 验证配置
func ValidateConfig(node fxjson.Node) error {
    if !node.Get("server").Exists() {
        return &ConfigError{"server", "服务器配置缺失"}
    }
    
    port := node.GetPath("server.port").IntOr(0)
    if port <= 0 || port > 65535 {
        return &ConfigError{"server.port", "端口号必须在1-65535之间"}
    }
    
    return nil
}
```

### 2. 环境变量集成

```go
func LoadConfigWithEnv(filepath string) *Config {
    node := fxjson.FromBytes([]byte("{}"))
    
    if data, err := os.ReadFile(filepath); err == nil {
        node = fxjson.FromBytes(data)
    }
    
    return &Config{
        Server: ServerConfig{
            Host: getEnvOrDefault("SERVER_HOST", 
                node.GetPath("server.host").StringOr("localhost")),
            Port: getEnvIntOrDefault("SERVER_PORT", 
                int(node.GetPath("server.port").IntOr(8080))),
        },
        // ... 其他配置
    }
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}
```

### 3. 日志集成

```go
import (
    "log"
    "github.com/icloudza/fxjson"
)

func LoadConfigWithLogging(filepath string) *Config {
    log.Printf("加载配置文件: %s", filepath)
    
    data, err := os.ReadFile(filepath)
    if err != nil {
        log.Printf("读取配置文件失败: %v，使用默认配置", err)
        return getDefaultConfig()
    }
    
    node := fxjson.FromBytes(data)
    
    // 记录重要配置项
    log.Printf("服务器端口: %d", node.GetPath("server.port").IntOr(8080))
    log.Printf("数据库主机: %s", node.GetPath("database.host").StringOr("localhost"))
    
    var config Config
    if err := node.DecodeStruct(&config); err != nil {
        log.Printf("解析配置失败: %v，使用默认配置", err)
        return getDefaultConfig()
    }
    
    log.Println("配置加载成功")
    return &config
}
```

## 常见问题

### Q: 如何更新到最新版本？

```bash
go get -u github.com/icloudza/fxjson
go mod tidy
```

### Q: 如何检查当前版本？

```bash
go list -m github.com/icloudza/fxjson
```

### Q: 如何解决版本冲突？

如果遇到版本冲突，可以使用 `go mod why` 查看依赖关系：

```bash
go mod why github.com/icloudza/fxjson
```

### Q: 支持哪些 Go 版本？

FxJSON 支持 Go 1.18 及以上版本。建议使用最新的稳定版本以获得最佳性能。

## 下一步

安装完成后，建议：

1. 阅读 [基础概念](/guide/concepts) 了解核心概念 
2. 探索 [示例代码](/examples/) 获取实际应用灵感