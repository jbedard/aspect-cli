/*
Copyright © 2021 Aspect Build Systems Inc

Not licensed for re-use.
*/

package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/natefinch/lumberjack"
)

type LoggerKeyType bool

const LoggerKey LoggerKeyType = true

type LoggerStruct struct {
	sample string
	l      *log.Logger
}

func (l *LoggerStruct) Test() {
	fmt.Println("testing 123")
}

var Logger LoggerStruct

func init() {
	Logger = LoggerStruct{
		sample: "testssssss",
	}

	// How do we get the invocation ID into here......
	// Make sure to use a timestamp so that the folders are ordered correctly
	// time.Now().UTC().Nano()
	// time.Now().UTC().UnixNano()
	e, err := os.OpenFile("/Users/jesse/Development/aspect-cli/foo.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		fmt.Printf("error opening file: %v", err)
		os.Exit(1)
	}

	Logger.l = log.New(e, "", log.Ldate|log.Ltime)
	Logger.l.SetOutput(&lumberjack.Logger{
		Filename:   "/Users/jesse/Development/aspect-cli/foo.log",
		MaxSize:    1,  // megabytes after which new file is created
		MaxBackups: 3,  // number of backups
		MaxAge:     28, //days
	})

}

func Test1() {
	Logger.Test()
	fmt.Println(Logger.sample)
	fmt.Println("testing 1234")
}

func Log(message string) {
	Logger.l.Println(message)

	// if log level = debug?????
	// Will tell the user where the logs were called from
	if false {
		pc, _, _, ok := runtime.Caller(1)
		details := runtime.FuncForPC(pc)
		if ok && details != nil {
			fmt.Printf("called from %s\n", details.Name())
		}
	}
}
