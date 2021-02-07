package logging

import (
	"fmt"
	"log"
	"os"
)

type Level int

var (
	F *os.File

	DefaultPrefix      = ""
	DefaultCallerDepth = 2

	logger     *log.Logger
	logPrefix  = ""
	levelFlags = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
)

const (
	DEBUG Level = iota
	INFO
	WARNING
	ERROR
	FATAL
)

func Init() {
	filePath := getLogFileFullPath()
	F = openLogFile(filePath)

	logger = log.New(F, DefaultPrefix, log.LstdFlags)
}

func Debug(v ...interface{}) {
	Init()
	setPrefix(DEBUG)
	logger.Println(v)
}

func Info(v ...interface{}) {
	Init()
	setPrefix(INFO)
	logger.Println(v)
}

func Warn(v ...interface{}) {
	Init()
	setPrefix(WARNING)
	logger.Println(v)
}

func Error(v ...interface{}) {
	Init()
	setPrefix(ERROR)
	logger.Println(v)
}

func Fatal(v ...interface{}) {
	Init()
	setPrefix(FATAL)
	logger.Fatalln(v)
}

func setPrefix(level Level) {
	Init()
	logPrefix = fmt.Sprintf("[%s]", levelFlags[level])
	logger.SetPrefix(logPrefix)
}
