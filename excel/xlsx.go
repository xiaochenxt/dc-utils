package excel

import (
	"errors"
	"fmt"
	"github.com/xuri/excelize/v2"
	"io"
	"os"
	"strings"
)

// XLSXStreamWriter 用于流式生成XLSX文件
type XLSXStreamWriter struct {
	f            *excelize.File
	streamWriter *excelize.StreamWriter
	sheetName    string
	output       io.Writer
	filePath     string
	file         *os.File
	rowIndex     int
	closed       bool
}

// NewXLSXStreamWriter 创建输出到io.Writer的XLSX写入器
func NewXLSXStreamWriter(output io.Writer, sheetName string) (*XLSXStreamWriter, error) {
	if output == nil {
		return nil, errors.New("输出io.Writer不能为空")
	}
	f := excelize.NewFile()
	// 验证并清理工作表名称
	cleanName, err := validateSheetName(sheetName)
	if err != nil {
		if closeErr := f.Close(); closeErr != nil {
			return nil, fmt.Errorf("创建XLSX写入器失败: %w; 关闭文件时发生额外错误: %v", err, closeErr)
		}
		return nil, err
	}
	// 创建工作表
	if cleanName != "Sheet1" {
		if _, err := f.NewSheet(cleanName); err != nil {
			if closeErr := f.Close(); closeErr != nil {
				return nil, fmt.Errorf("创建工作表失败: %w; 关闭文件时发生额外错误: %v", err, closeErr)
			}
			return nil, err
		}
		sheetIndex, err := f.GetSheetIndex(cleanName)
		if err != nil {
			if closeErr := f.Close(); closeErr != nil {
				return nil, fmt.Errorf("获取工作表索引失败: %w; 关闭文件时发生额外错误: %v", err, closeErr)
			}
			return nil, err
		}
		f.SetActiveSheet(sheetIndex)
		if err := f.DeleteSheet("Sheet1"); err != nil {
			if closeErr := f.Close(); closeErr != nil {
				return nil, fmt.Errorf("删除默认工作表失败: %w; 关闭文件时发生额外错误: %v", err, closeErr)
			}
			return nil, err
		}
	}
	// 创建流式写入器
	streamWriter, err := f.NewStreamWriter(cleanName)
	if err != nil {
		if closeErr := f.Close(); closeErr != nil {
			return nil, fmt.Errorf("创建XLSX流式写入器失败: %w; 关闭文件时发生额外错误: %v", err, closeErr)
		}
		return nil, err
	}
	return &XLSXStreamWriter{
		f:            f,
		streamWriter: streamWriter,
		sheetName:    cleanName,
		output:       output,
		rowIndex:     1,
		closed:       false,
	}, nil
}

// NewXLSXStreamWriterForFile 创建输出到文件的XLSX写入器
func NewXLSXStreamWriterForFile(filePath, sheetName string) (*XLSXStreamWriter, error) {
	if filePath == "" {
		return nil, errors.New("文件路径不能为空")
	}
	f := excelize.NewFile()
	// 验证并清理工作表名称
	cleanName, err := validateSheetName(sheetName)
	if err != nil {
		if closeErr := f.Close(); closeErr != nil {
			return nil, fmt.Errorf("创建XLSX写入器失败: %w; 关闭文件时发生额外错误: %v", err, closeErr)
		}
		return nil, err
	}
	// 创建工作表
	if cleanName != "Sheet1" {
		if _, err := f.NewSheet(cleanName); err != nil {
			if closeErr := f.Close(); closeErr != nil {
				return nil, fmt.Errorf("创建工作表失败: %w; 关闭文件时发生额外错误: %v", err, closeErr)
			}
			return nil, err
		}
		sheetIndex, err := f.GetSheetIndex(cleanName)
		if err != nil {
			if closeErr := f.Close(); closeErr != nil {
				return nil, fmt.Errorf("获取工作表索引失败: %w; 关闭文件时发生额外错误: %v", err, closeErr)
			}
			return nil, err
		}
		f.SetActiveSheet(sheetIndex)
		if err := f.DeleteSheet("Sheet1"); err != nil {
			if closeErr := f.Close(); closeErr != nil {
				return nil, fmt.Errorf("删除默认工作表失败: %w; 关闭文件时发生额外错误: %v", err, closeErr)
			}
			return nil, err
		}
	}
	// 创建流式写入器
	streamWriter, err := f.NewStreamWriter(cleanName)
	if err != nil {
		if closeErr := f.Close(); closeErr != nil {
			return nil, fmt.Errorf("创建XLSX流式写入器失败: %w; 关闭文件时发生额外错误: %v", err, closeErr)
		}
		return nil, err
	}
	return &XLSXStreamWriter{
		f:            f,
		streamWriter: streamWriter,
		sheetName:    cleanName,
		filePath:     filePath,
		rowIndex:     1,
		closed:       false,
	}, nil
}

// validateSheetName 验证并清理工作表名称
func validateSheetName(name string) (string, error) {
	if name == "" {
		return "Sheet1", nil
	}
	// 移除非法字符
	cleanName := strings.Map(func(r rune) rune {
		switch r {
		case '\\', '/', ':', '*', '?', '"', '<', '>', '|':
			return -1 // 移除非法字符
		default:
			return r
		}
	}, name)
	// 检查长度（Excel限制为31个字符）
	if len(cleanName) > 31 {
		cleanName = cleanName[:31]
	}
	// 确保清理后的名称不为空
	if cleanName == "" {
		return "", fmt.Errorf("无效的工作表名称: %s", name)
	}
	return cleanName, nil
}

// WriteHeader 写入表头
func (w *XLSXStreamWriter) WriteHeader(header []any) error {
	if w.closed {
		return errors.New("写入器已关闭")
	}
	if header == nil {
		return errors.New("表头不能为空")
	}
	return w.writeRow(header)
}

// WriteRow 写入一行数据
func (w *XLSXStreamWriter) WriteRow(row []any) error {
	if w.closed {
		return errors.New("写入器已关闭")
	}
	if row == nil {
		return errors.New("行数据不能为空")
	}
	err := w.writeRow(row)
	if err != nil {
		return err
	}
	err = w.streamWriter.Flush()
	if err != nil {
		return err
	}
	return err
}

// WriteRows 写入多行数据
func (w *XLSXStreamWriter) WriteRows(rows [][]any) error {
	if w == nil {
		return errors.New("XLSXStreamWriter不能为空")
	}

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
	err := w.streamWriter.Flush()
	if err != nil {
		return err
	}
	return nil
}

// 内部写入行方法
func (w *XLSXStreamWriter) writeRow(row []any) error {
	if w.streamWriter == nil {
		return errors.New("XLSX流式写入器未初始化")
	}
	// 转换数据类型
	values := make([]any, len(row))
	for i, v := range row {
		values[i] = v
	}
	// 生成单元格坐标
	cell, err := excelize.CoordinatesToCellName(1, w.rowIndex)
	if err != nil {
		return fmt.Errorf("生成单元格坐标失败: %w", err)
	}
	// 写入行数据
	if err := w.streamWriter.SetRow(cell, values); err != nil {
		return fmt.Errorf("写入行失败: %w", err)
	}
	w.rowIndex++
	return nil
}

// Close 关闭写入器并完成输出
func (w *XLSXStreamWriter) Close() error {
	if w.closed {
		return nil
	}
	w.closed = true
	// 刷新流
	if w.streamWriter != nil {
		if err := w.streamWriter.Flush(); err != nil {
			return fmt.Errorf("刷新流失败: %w", err)
		}
	}
	// 输出到文件或writer
	var err error
	if w.output != nil {
		err = w.f.Write(w.output)
	} else if w.filePath != "" {
		err = w.f.SaveAs(w.filePath)
	}
	// 关闭excelize文件
	if closeErr := w.f.Close(); closeErr != nil {
		if err != nil {
			return fmt.Errorf("输出失败: %w; 关闭Excel文件时发生额外错误: %v", err, closeErr)
		}
		return fmt.Errorf("关闭Excel文件失败: %w", closeErr)
	}
	w.f = nil
	// 关闭文件（如果打开过）
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return fmt.Errorf("关闭文件失败: %w", err)
		}
		w.file = nil
	}
	return err
}
