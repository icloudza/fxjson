package fxjson

import (
	"fmt"
	"reflect"
	"sync"
)

// BatchMarshaler 批量序列化器
type BatchMarshaler struct {
	buf     *Buffer
	opts    SerializeOptions
	workers int
}

// NewBatchMarshaler 创建批量序列化器
func NewBatchMarshaler(opts SerializeOptions, workers int) *BatchMarshaler {
	if workers <= 0 {
		workers = 1
	}
	return &BatchMarshaler{
		buf:     getBuffer(),
		opts:    opts,
		workers: workers,
	}
}

// Close 关闭批量序列化器，释放资源
func (bm *BatchMarshaler) Close() {
	if bm.buf != nil {
		putBuffer(bm.buf)
		bm.buf = nil
	}
}

// MarshalSlice 批量序列化切片
func (bm *BatchMarshaler) MarshalSlice(slice interface{}) ([]byte, error) {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, fmt.Errorf("expected slice or array, got %s", rv.Kind())
	}

	length := rv.Len()
	if length == 0 {
		return []byte("[]"), nil
	}

	// 单线程处理小切片
	if length < 100 || bm.workers == 1 {
		return bm.marshalSliceSequential(rv)
	}

	// 多线程处理大切片
	return bm.marshalSliceConcurrent(rv)
}

// marshalSliceSequential 顺序序列化切片
func (bm *BatchMarshaler) marshalSliceSequential(rv reflect.Value) ([]byte, error) {
	bm.buf.Reset()
	bm.buf.WriteByte('[')

	length := rv.Len()
	for i := 0; i < length; i++ {
		if i > 0 {
			bm.buf.WriteByte(',')
		}

		if bm.opts.Indent != "" {
			bm.buf.WriteByte('\n')
			writeIndent(bm.buf, bm.opts.Indent, 1)
		}

		if err := marshalValue(bm.buf, rv.Index(i), bm.opts, 1); err != nil {
			return nil, err
		}
	}

	if bm.opts.Indent != "" && length > 0 {
		bm.buf.WriteByte('\n')
	}

	bm.buf.WriteByte(']')

	result := make([]byte, len(bm.buf.buf))
	copy(result, bm.buf.buf)
	return result, nil
}

// marshalSliceConcurrent 并发序列化切片
func (bm *BatchMarshaler) marshalSliceConcurrent(rv reflect.Value) ([]byte, error) {
	length := rv.Len()
	chunkSize := (length + bm.workers - 1) / bm.workers

	type chunkResult struct {
		index int
		data  []byte
		err   error
	}

	results := make(chan chunkResult, bm.workers)
	var wg sync.WaitGroup

	// 启动工作协程
	for i := 0; i < bm.workers; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > length {
			end = length
		}
		if start >= length {
			break
		}

		wg.Add(1)
		go func(chunkIndex, chunkStart, chunkEnd int) {
			defer wg.Done()

			buf := getBuffer()
			defer putBuffer(buf)

			for j := chunkStart; j < chunkEnd; j++ {
				if j > chunkStart {
					buf.WriteByte(',')
				}

				if err := marshalValue(buf, rv.Index(j), bm.opts, 0); err != nil {
					results <- chunkResult{index: chunkIndex, err: err}
					return
				}
			}

			result := make([]byte, len(buf.buf))
			copy(result, buf.buf)
			results <- chunkResult{index: chunkIndex, data: result}
		}(i, start, end)
	}

	// 等待所有协程完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	chunks := make([][]byte, bm.workers)
	for result := range results {
		if result.err != nil {
			return nil, result.err
		}
		chunks[result.index] = result.data
	}

	// 合并结果
	bm.buf.Reset()
	bm.buf.WriteByte('[')

	for i, chunk := range chunks {
		if chunk != nil {
			if i > 0 {
				bm.buf.WriteByte(',')
			}
			bm.buf.Write(chunk)
		}
	}

	bm.buf.WriteByte(']')

	result := make([]byte, len(bm.buf.buf))
	copy(result, bm.buf.buf)
	return result, nil
}

// BatchMarshalStructs 批量序列化多个结构体
func BatchMarshalStructs(structs []interface{}) ([][]byte, error) {
	return BatchMarshalStructsWithOptions(structs, DefaultSerializeOptions)
}

// BatchMarshalStructsWithOptions 使用指定选项批量序列化多个结构体
func BatchMarshalStructsWithOptions(structs []interface{}, opts SerializeOptions) ([][]byte, error) {
	if len(structs) == 0 {
		return nil, nil
	}

	results := make([][]byte, len(structs))

	for i, s := range structs {
		data, err := MarshalWithOptions(s, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal struct at index %d: %v", i, err)
		}
		results[i] = data
	}

	return results, nil
}

// BatchMarshalStructsConcurrent 并发批量序列化多个结构体
func BatchMarshalStructsConcurrent(structs []interface{}, workers int) ([][]byte, error) {
	return BatchMarshalStructsConcurrentWithOptions(structs, DefaultSerializeOptions, workers)
}

// BatchMarshalStructsConcurrentWithOptions 使用指定选项并发批量序列化多个结构体
func BatchMarshalStructsConcurrentWithOptions(structs []interface{}, opts SerializeOptions, workers int) ([][]byte, error) {
	if len(structs) == 0 {
		return nil, nil
	}

	if workers <= 0 {
		workers = 1
	}

	if len(structs) < workers {
		workers = len(structs)
	}

	type task struct {
		index int
		value interface{}
	}

	type result struct {
		index int
		data  []byte
		err   error
	}

	tasks := make(chan task, len(structs))
	results := make(chan result, len(structs))

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {
				data, err := MarshalWithOptions(task.value, opts)
				results <- result{
					index: task.index,
					data:  data,
					err:   err,
				}
			}
		}()
	}

	// 发送任务
	for i, s := range structs {
		tasks <- task{index: i, value: s}
	}
	close(tasks)

	// 等待完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	output := make([][]byte, len(structs))
	for result := range results {
		if result.err != nil {
			return nil, fmt.Errorf("failed to marshal struct at index %d: %v", result.index, result.err)
		}
		output[result.index] = result.data
	}

	return output, nil
}

// StreamMarshaler 流式序列化器（大数据处理）
type StreamMarshaler struct {
	writer   func([]byte) error
	opts     SerializeOptions
	first    bool
	inArray  bool
	inObject bool
	closed   bool
}

// NewStreamMarshaler 创建流式序列化器
func NewStreamMarshaler(writer func([]byte) error, opts SerializeOptions) *StreamMarshaler {
	return &StreamMarshaler{
		writer: writer,
		opts:   opts,
		first:  true,
	}
}

// StartArray 开始数组序列化
func (sm *StreamMarshaler) StartArray() error {
	if sm.closed {
		return fmt.Errorf("marshaler is closed")
	}

	sm.inArray = true
	sm.first = true
	return sm.writer([]byte{'['})
}

// EndArray 结束数组序列化
func (sm *StreamMarshaler) EndArray() error {
	if sm.closed {
		return fmt.Errorf("marshaler is closed")
	}

	sm.inArray = false
	return sm.writer([]byte{']'})
}

// StartObject 开始对象序列化
func (sm *StreamMarshaler) StartObject() error {
	if sm.closed {
		return fmt.Errorf("marshaler is closed")
	}

	sm.inObject = true
	sm.first = true
	return sm.writer([]byte{'{'})
}

// EndObject 结束对象序列化
func (sm *StreamMarshaler) EndObject() error {
	if sm.closed {
		return fmt.Errorf("marshaler is closed")
	}

	sm.inObject = false
	return sm.writer([]byte{'}'})
}

// WriteValue 写入值
func (sm *StreamMarshaler) WriteValue(v interface{}) error {
	if sm.closed {
		return fmt.Errorf("marshaler is closed")
	}

	if sm.inArray || sm.inObject {
		if !sm.first {
			if err := sm.writer([]byte{','}); err != nil {
				return err
			}
		}
		sm.first = false
	}

	data, err := MarshalWithOptions(v, sm.opts)
	if err != nil {
		return err
	}

	return sm.writer(data)
}

// WriteField 写入对象字段（键值对）
func (sm *StreamMarshaler) WriteField(key string, value interface{}) error {
	if sm.closed {
		return fmt.Errorf("marshaler is closed")
	}

	if !sm.inObject {
		return fmt.Errorf("not in object context")
	}

	if !sm.first {
		if err := sm.writer([]byte{','}); err != nil {
			return err
		}
	}
	sm.first = false

	// 写入键
	buf := getBuffer()
	defer putBuffer(buf)

	writeString(buf, key, sm.opts.EscapeHTML)
	buf.WriteByte(':')

	if sm.opts.Indent != "" {
		buf.WriteByte(' ')
	}

	if err := sm.writer(buf.Bytes()); err != nil {
		return err
	}

	// 写入值
	data, err := MarshalWithOptions(value, sm.opts)
	if err != nil {
		return err
	}

	return sm.writer(data)
}

// Close 关闭流式序列化器
func (sm *StreamMarshaler) Close() error {
	sm.closed = true
	return nil
}

// MarshalToWriter 将数据序列化到writer
func MarshalToWriter(v interface{}, writer func([]byte) error) error {
	return MarshalToWriterWithOptions(v, writer, DefaultSerializeOptions)
}

// MarshalToWriterWithOptions 使用指定选项将数据序列化到writer
func MarshalToWriterWithOptions(v interface{}, writer func([]byte) error, opts SerializeOptions) error {
	data, err := MarshalWithOptions(v, opts)
	if err != nil {
		return err
	}
	return writer(data)
}

// ChunkedMarshal 分块序列化大对象
func ChunkedMarshal(v interface{}, chunkSize int) ([][]byte, error) {
	data, err := Marshal(v)
	if err != nil {
		return nil, err
	}

	if chunkSize <= 0 {
		chunkSize = 1024
	}

	var chunks [][]byte
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}

		chunk := make([]byte, end-i)
		copy(chunk, data[i:end])
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}
