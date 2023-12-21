package tos

import "fmt"

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

var stdlog stdLogger

type stdLogger struct {
}

func (s stdLogger) Debug(args ...interface{}) {
	fmt.Println(args...)
}

func (s stdLogger) Info(args ...interface{}) {
	fmt.Println(args...)

}

func (s stdLogger) Warn(args ...interface{}) {
	fmt.Println(args...)

}

func (s stdLogger) Error(args ...interface{}) {
	fmt.Println(args...)

}

func (s stdLogger) Fatal(args ...interface{}) {
	fmt.Println(args...)
}
