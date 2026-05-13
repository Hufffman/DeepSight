package main

import (
	"DeepSight/internal/app"
	"DeepSight/internal/util/logger"

	"go.uber.org/zap"
)

func main() {
	if err := app.Run(); err != nil {
		logger.Log.Fatal("application exited", zap.Error(err))
	}
}
