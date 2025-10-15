package excel

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

// CSVStreamWriter 用于流式生成CSV文件
type CSVStreamWriter struct {
	csvWriter *csv.Writer
	output    io.Writer
	filePath  string
	file      *os.File
	delimiter rune
	rowIndex  int
	closed    bool
}

// NewCSVStreamWriter 创建输出到io.Writer的CSV写入器
func NewCSVStreamWriter(output io.Writer, delimiter rune) (*CSVStreamWriter, error) {
	if output == nil {
		return nil, errors.New("输出io.Writer不能为空")
	}
	// 设置默认分隔符
	if delimiter == 0 {
		delimiter = ','
	}
	// 创建CSV写入器
	csvWriter := csv.NewWriter(output)
	csvWriter.Comma = delimiter
	return &CSVStreamWriter{
		csvWriter: csvWriter,
		output:    output,
		delimiter: delimiter,
		rowIndex:  1,
		closed:    false,
	}, nil
}

// NewCSVStreamWriterForFile 创建输出到文件的CSV写入器
func NewCSVStreamWriterForFile(filePath string, delimiter rune) (*CSVStreamWriter, error) {
	if filePath == "" {
		return nil, errors.New("文件路径不能为空")
	}
	// 打开文件
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("创建CSV文件失败: %w", err)
	}
	// 设置默认分隔符
	if delimiter == 0 {
		delimiter = ','
	}
	// 创建CSV写入器
	csvWriter := csv.NewWriter(file)
	csvWriter.Comma = delimiter
	return &CSVStreamWriter{
		csvWriter: csvWriter,
		filePath:  filePath,
		file:      file,
		delimiter: delimiter,
		rowIndex:  1,
		closed:    false,
	}, nil
}

// WriteHeader 写入表头
func (w *CSVStreamWriter) WriteHeader(header []any) error {
	if w.closed {
		return errors.New("写入器已关闭")
	}
	if header == nil {
		return errors.New("表头不能为空")
	}
	return w.writeRow(header)
}

// WriteRow 写入一行数据
func (w *CSVStreamWriter) WriteRow(row []any) error {
	if w.closed {
		return nil
	}
	if row == nil {
		return nil
	}
	err := w.writeRow(row)
	if err != nil {
		return err
	}
	w.csvWriter.Flush()
	return err
}

// WriteRows 写入多行数据
func (w *CSVStreamWriter) WriteRows(rows [][]any) error {
	if w.closed {
		return errors.New("写入器已关闭")
	}
	if rows == nil {
		return errors.New("多行数据不能为空")
	}
	for i, row := range rows {
		if row == nil {
			return fmt.Errorf("第 %d 行数据为空", i+1)
		}

		if err := w.writeRow(row); err != nil {
			return fmt.Errorf("写入第 %d 行失败: %w", i+1, err)
		}
	}
	w.csvWriter.Flush()
	return nil
}

// 内部写入行方法
func (w *CSVStreamWriter) writeRow(row []any) error {
	// 转换为字符串切片
	stringRow := make([]string, len(row))
	for i, v := range row {
		stringRow[i] = formatValue(v)
	}
	// 写入行数据
	if err := w.csvWriter.Write(stringRow); err != nil {
		return fmt.Errorf("写入CSV行失败: %w", err)
	}
	w.rowIndex++
	return nil
}

// Close 关闭写入器并完成输出
func (w *CSVStreamWriter) Close() error {
	if w.closed {
		return nil
	}
	w.closed = true
	// 刷新缓冲区
	if w.csvWriter != nil {
		w.csvWriter.Flush()
		if err := w.csvWriter.Error(); err != nil {
			return fmt.Errorf("刷新CSV缓冲区失败: %w", err)
		}
	}

	// 关闭文件（如果打开过）
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return fmt.Errorf("关闭CSV文件失败: %w", err)
		}
		w.file = nil
	}

	return nil
}

// formatValue 将各种类型的值转换为字符串
func formatValue(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}
