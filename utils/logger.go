/**
  @author: Bruce
  @since: 2023/5/19
  @desc: //日志系统
**/

package utils

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type LogLevel uint16

// 定义日志级别
const (
	UNKNOWN LogLevel = iota // 0
	DEBUG
	TRACE
	INFO
	WARNING
	ERROR
	FATAL
)

var LogFileName string
var Log Logger

// Logger 日志结构体
type Logger struct {
	Level LogLevel
}

func parseLogLevel(s string) (LogLevel, error) {
	s = strings.ToLower(s)
	switch s {
	case "debug":
		return DEBUG, nil
	case "trace":
		return TRACE, nil
	case "info":
		return INFO, nil
	case "warning":
		return WARNING, nil
	case "error":
		return ERROR, nil
	case "fatal":
		return FATAL, nil
	default:
		err := errors.New("无效的日志级别")
		return UNKNOWN, err
	}

}

func getLogLevelStr(logLevel LogLevel) string {
	switch logLevel {
	case DEBUG:
		return "debug"
	case TRACE:
		return "trace"
	case INFO:
		return "info"
	case WARNING:
		return "warning"
	case ERROR:
		return "error"
	case FATAL:
		return "fatal"
	default:
		return "unknown"
	}
}

// 获取函数名、文件名、行号
// skip表示隔了几层
func getInfo(skip int) (funcName string, fileName string, lineNo int) {
	// pc：函数信息
	// file：文件
	// line：行号，也就是当前行号
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		fmt.Printf("runtime.Caller() failed")
		return
	}
	funName := runtime.FuncForPC(pc).Name()

	return funName, path.Base(file), line
}

// 判断啥级别的日志可以输出是否输出
func (l Logger) enable(logLevel LogLevel) bool {
	return logLevel >= l.Level
}

func (l Logger) Debug(msg string) {
	if l.enable(DEBUG) {
		printLog(DEBUG, msg)
	}
}

func (l Logger) TRACE(msg string) {
	if l.enable(TRACE) {
		printLog(TRACE, msg)
	}
}

func (l Logger) Info(msg string) {
	if l.enable(INFO) {
		printLog(INFO, msg)
	}
}

func (l Logger) Warning(msg string) {

	if l.enable(WARNING) {
		printLog(WARNING, msg)
	}
}

func (l Logger) Error(msg string) {
	if l.enable(ERROR) {
		printLog(ERROR, msg)
	}
}

func (l Logger) Fatal(msg string) {
	if l.enable(FATAL) {
		printLog(FATAL, msg)
	}
}

func printLog(lv LogLevel, msg string) {
	now := time.Now().Format("2006-01-02 15:04:05")
	// 拿到第二层的函数名
	funcName, filePath, lineNo := getInfo(3)
	log := fmt.Sprintf("[%s] [%s] [%s:%s:%d] %s", now, getLogLevelStr(lv), filePath, funcName, lineNo, msg)
	fmt.Println(log)
	if LogFileName == "" {
		return
	}
	logFile, err := os.OpenFile(LogFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("open log file failed, err:", err)
		return
	}

	fprintf, err := fmt.Fprintln(logFile, log)
	if err != nil {
		fmt.Println(fprintf, "write log file failed, err:", err)
		return
	}

}

func InitLoggerDefault() (Logger, error) {
	level := Conf.LogLevel
	logPath := Conf.LogPath
	Log = NewLog(level)
	LogFileName = getFileName(logPath, "log")
	return Log, nil
}

func InitLogger(level string, path ...string) (Logger, error) {
	Log = NewLog(level)
	if path == nil {
		LogFileName = ""
		return Log, nil
	}
	LogFileName = getFileName(path[0], "log")
	return Log, nil
}

// NewLog Logger构造方法
func NewLog(levelStr string) Logger {
	level, err := parseLogLevel(levelStr)
	if err != nil {
		panic(err)
	}
	// 构造了一个Logger对象
	return Logger{
		Level: level,
	}
}

func getFileName(basePath, name string) string {
	timeObj := time.Now()
	year := timeObj.Year()
	month := timeObj.Month()
	day := timeObj.Day()
	logName := fmt.Sprintf("%d-%02d-%02d-%s", year, month, day, name)
	folderName := fmt.Sprintf("%d-%02d", year, month)
	folderPath := filepath.Join(basePath, folderName)
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		// 必须分成两步
		// 先创建文件夹
		err := os.Mkdir(folderPath, 0777)
		if err != nil {
			println("getFileName: 创建文件夹失败")
		}
		// 再修改权限
		err = os.Chmod(folderPath, 0777)
		if err != nil {
			println("getFileName: 修改文件夹权限失败")
		}
	}
	logName = filepath.Join(folderPath, logName)
	return logName
}
