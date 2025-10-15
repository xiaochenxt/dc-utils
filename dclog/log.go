package dclog

import (
	"context"
	"fmt"
	fiberLog "github.com/gofiber/fiber/v2/log"
	"io"
	"log"
	"os"
	"strings"
)

// 默认 logger 实例
var logger fiberLog.AllLogger = &dcLogger{
	stdlog: log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds), // 移除了log.Lshortfile标志
	depth:  4,
	level:  fiberLog.LevelInfo,
}

type dcLogger struct {
	stdlog *log.Logger
	depth  int
	level  fiberLog.Level
}

func (l *dcLogger) Trace(v ...any) {
	if l.level > fiberLog.LevelTrace {
		return
	}
	_ = l.stdlog.Output(l.depth, fmt.Sprint("[Trace] ", fmt.Sprint(v...)))
}

func (l *dcLogger) Debug(v ...any) {
	if l.level > fiberLog.LevelDebug {
		return
	}
	_ = l.stdlog.Output(l.depth, fmt.Sprint("[Debug] ", fmt.Sprint(v...)))
}

func (l *dcLogger) Info(v ...any) {
	if l.level > fiberLog.LevelInfo {
		return
	}
	_ = l.stdlog.Output(l.depth, fmt.Sprint("[Info] ", fmt.Sprint(v...)))
}

func (l *dcLogger) Warn(v ...any) {
	if l.level > fiberLog.LevelWarn {
		return
	}
	_ = l.stdlog.Output(l.depth, fmt.Sprint("[Warn] ", fmt.Sprint(v...)))
}

func (l *dcLogger) Error(v ...any) {
	if l.level > fiberLog.LevelError {
		return
	}
	_ = l.stdlog.Output(l.depth, fmt.Sprint("[Error] ", fmt.Sprint(v...)))
}

func (l *dcLogger) Fatal(v ...any) {
	_ = l.stdlog.Output(l.depth, fmt.Sprint("[Fatal] ", fmt.Sprint(v...)))
	os.Exit(1)
}

func (l *dcLogger) Panic(v ...any) {
	s := fmt.Sprint("[Panic] ", fmt.Sprint(v...))
	_ = l.stdlog.Output(l.depth, s)
	panic(s)
}

func (l *dcLogger) Tracef(format string, v ...any) {
	if l.level > fiberLog.LevelTrace {
		return
	}
	_ = l.stdlog.Output(l.depth, fmt.Sprintf("[Trace] "+format, v...))
}

func (l *dcLogger) Debugf(format string, v ...any) {
	if l.level > fiberLog.LevelDebug {
		return
	}
	_ = l.stdlog.Output(l.depth, fmt.Sprintf("[Debug] "+format, v...))
}

func (l *dcLogger) Infof(format string, v ...any) {
	if l.level > fiberLog.LevelInfo {
		return
	}
	_ = l.stdlog.Output(l.depth, fmt.Sprintf("[Info] "+format, v...))
}

func (l *dcLogger) Warnf(format string, v ...any) {
	if l.level > fiberLog.LevelWarn {
		return
	}
	_ = l.stdlog.Output(l.depth, fmt.Sprintf("[Warn] "+format, v...))
}

func (l *dcLogger) Errorf(format string, v ...any) {
	if l.level > fiberLog.LevelError {
		return
	}
	_ = l.stdlog.Output(l.depth, fmt.Sprintf("[Error] "+format, v...))
}

func (l *dcLogger) Fatalf(format string, v ...any) {
	_ = l.stdlog.Output(l.depth, fmt.Sprintf("[Fatal] "+format, v...))
	os.Exit(1)
}

func (l *dcLogger) Panicf(format string, v ...any) {
	s := fmt.Sprintf("[Panic] "+format, v...)
	_ = l.stdlog.Output(l.depth, s)
	panic(s)
}

func (l *dcLogger) Tracew(msg string, keysAndValues ...any) {
	if l.level > fiberLog.LevelTrace {
		return
	}
	_ = l.stdlog.Output(l.depth, formatWithKV("[Trace] ", msg, keysAndValues...))
}

func (l *dcLogger) Debugw(msg string, keysAndValues ...any) {
	if l.level > fiberLog.LevelDebug {
		return
	}
	_ = l.stdlog.Output(l.depth, formatWithKV("[Debug] ", msg, keysAndValues...))
}

func (l *dcLogger) Infow(msg string, keysAndValues ...any) {
	if l.level > fiberLog.LevelInfo {
		return
	}
	_ = l.stdlog.Output(l.depth, formatWithKV("[Info] ", msg, keysAndValues...))
}

func (l *dcLogger) Warnw(msg string, keysAndValues ...any) {
	if l.level > fiberLog.LevelWarn {
		return
	}
	_ = l.stdlog.Output(l.depth, formatWithKV("[Warn] ", msg, keysAndValues...))
}

func (l *dcLogger) Errorw(msg string, keysAndValues ...any) {
	if l.level > fiberLog.LevelError {
		return
	}
	_ = l.stdlog.Output(l.depth, formatWithKV("[Error] ", msg, keysAndValues...))
}

func (l *dcLogger) Fatalw(msg string, keysAndValues ...any) {
	_ = l.stdlog.Output(l.depth, formatWithKV("[Fatal] ", msg, keysAndValues...))
	os.Exit(1)
}

func (l *dcLogger) Panicw(msg string, keysAndValues ...any) {
	s := formatWithKV("[Panic] ", msg, keysAndValues...)
	_ = l.stdlog.Output(l.depth, s)
	panic(s)
}

// 格式化键值对日志
func formatWithKV(prefix, msg string, keysAndValues ...any) string {
	if len(keysAndValues) == 0 {
		return prefix + msg
	}

	// 确保键值对是偶数个
	if len(keysAndValues)%2 != 0 {
		keysAndValues = append(keysAndValues, "MISSING_VALUE")
	}

	var builder strings.Builder
	builder.WriteString(prefix)
	builder.WriteString(msg)

	for i := 0; i < len(keysAndValues); i += 2 {
		if i == 0 {
			builder.WriteString(" | ")
		} else {
			builder.WriteString(" ")
		}
		builder.WriteString(fmt.Sprintf("%v=%v", keysAndValues[i], keysAndValues[i+1]))
	}

	return builder.String()
}

func (l *dcLogger) SetLevel(level fiberLog.Level) {
	l.level = level
}

func (l *dcLogger) SetOutput(writer io.Writer) {
	l.stdlog.SetOutput(writer)
}

func (l *dcLogger) WithContext(ctx context.Context) fiberLog.CommonLogger {
	// 注意：这里简单返回原 logger，实际应用中可以从 ctx 中提取元数据
	return &dcLogger{
		stdlog: l.stdlog,
		depth:  l.depth + 1, // 增加调用栈深度
		level:  l.level,
	}
}

func DefaultLogger() fiberLog.AllLogger {
	return logger
}

// NewLogger 创建一个新的 logger 实例
func NewLogger(output io.Writer, level fiberLog.Level) fiberLog.AllLogger {
	return &dcLogger{
		stdlog: log.New(output, "", log.LstdFlags|log.Lmicroseconds), // 移除了log.Lshortfile标志
		depth:  4,
		level:  level,
	}
}
