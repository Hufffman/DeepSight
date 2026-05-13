package logger

import (
	"strings"

	"go.uber.org/zap"
)

var Log = zap.Must(zap.NewDevelopment())

func Init(env string) error {
	var (
		l   *zap.Logger
		err error
	)

	if strings.EqualFold(env, "production") {
		l, err = zap.NewProduction()
	} else {
		l, err = zap.NewDevelopment()
	}
	if err != nil {
		return err
	}

	if err := Log.Sync(); err != nil {
		_ = err
	}

	Log = l
	zap.ReplaceGlobals(l)
	return nil
}

func L() *zap.Logger {
	return Log
}

func Sync() error {
	return Log.Sync()
}
