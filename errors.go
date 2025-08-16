package fxjson

import (
	"fmt"
)

// ErrorType 错误类型
type ErrorType int

const (
	// ErrorTypeInvalidJSON 无效的JSON格式
	ErrorTypeInvalidJSON ErrorType = iota
	// ErrorTypeOutOfBounds 越界错误
	ErrorTypeOutOfBounds
	// ErrorTypeTypeMismatch 类型不匹配
	ErrorTypeTypeMismatch
	// ErrorTypeMemoryLimit 内存限制
	ErrorTypeMemoryLimit
	// ErrorTypeDepthLimit 深度限制
	ErrorTypeDepthLimit
	// ErrorTypeNotFound 未找到
	ErrorTypeNotFound
	// ErrorTypeValidation 验证错误
	ErrorTypeValidation
)

// String 返回错误类型的字符串表示
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeInvalidJSON:
		return "InvalidJSON"
	case ErrorTypeOutOfBounds:
		return "OutOfBounds"
	case ErrorTypeTypeMismatch:
		return "TypeMismatch"
	case ErrorTypeMemoryLimit:
		return "MemoryLimit"
	case ErrorTypeDepthLimit:
		return "DepthLimit"
	case ErrorTypeNotFound:
		return "NotFound"
	case ErrorTypeValidation:
		return "Validation"
	default:
		return "Unknown"
	}
}

// FxJSONError FxJSON错误结构
type FxJSONError struct {
	Type    ErrorType
	Message string
	Context string
	Pos     int
	Line    int
	Column  int
	Cause   error
}

// Error 实现error接口
func (e *FxJSONError) Error() string {
	if e.Line > 0 && e.Column > 0 {
		return fmt.Sprintf("[%s] %s at line %d, column %d", e.Type, e.Message, e.Line, e.Column)
	}
	if e.Pos > 0 {
		return fmt.Sprintf("[%s] %s at position %d", e.Type, e.Message, e.Pos)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap 返回包装的错误
func (e *FxJSONError) Unwrap() error {
	return e.Cause
}

// Position 表示JSON中的位置
type Position struct {
	Offset int
	Line   int
	Column int
}

// CalculatePosition 计算给定偏移量在数据中的行列位置
func CalculatePosition(data []byte, offset int) Position {
	if offset < 0 || offset > len(data) {
		return Position{Offset: offset, Line: 0, Column: 0}
	}

	line, column := 1, 1
	for i := 0; i < offset && i < len(data); i++ {
		if data[i] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return Position{Offset: offset, Line: line, Column: column}
}

// NewContextError 创建带上下文的错误
func NewContextError(errorType ErrorType, message string, data []byte, pos int) *FxJSONError {
	position := CalculatePosition(data, pos)

	// 提取错误附近的上下文（前后20个字符）
	contextStart := max(0, pos-20)
	contextEnd := min(len(data), pos+20)
	context := ""
	if contextStart < contextEnd && contextEnd <= len(data) {
		context = string(data[contextStart:contextEnd])
	}

	return &FxJSONError{
		Type:    errorType,
		Message: message,
		Context: context,
		Pos:     pos,
		Line:    position.Line,
		Column:  position.Column,
	}
}

// NewTypeMismatchError 创建类型不匹配错误
func NewTypeMismatchError(expected, actual string, node Node) *FxJSONError {
	return &FxJSONError{
		Type:    ErrorTypeTypeMismatch,
		Message: fmt.Sprintf("expected %s, got %s", expected, actual),
		Context: string(node.Raw()),
	}
}

// NewNotFoundError 创建未找到错误
func NewNotFoundError(key string) *FxJSONError {
	return &FxJSONError{
		Type:    ErrorTypeNotFound,
		Message: fmt.Sprintf("key '%s' not found", key),
	}
}

// NewValidationError 创建验证错误
func NewValidationError(field, message string) *FxJSONError {
	return &FxJSONError{
		Type:    ErrorTypeValidation,
		Message: fmt.Sprintf("validation failed for field '%s': %s", field, message),
	}
}

// NewDepthLimitError 创建深度限制错误
func NewDepthLimitError(maxDepth, currentDepth int) *FxJSONError {
	return &FxJSONError{
		Type:    ErrorTypeDepthLimit,
		Message: fmt.Sprintf("maximum depth %d exceeded, current depth: %d", maxDepth, currentDepth),
	}
}

// NewMemoryLimitError 创建内存限制错误
func NewMemoryLimitError(limit, requested int) *FxJSONError {
	return &FxJSONError{
		Type:    ErrorTypeMemoryLimit,
		Message: fmt.Sprintf("memory limit %d exceeded, requested: %d", limit, requested),
	}
}
