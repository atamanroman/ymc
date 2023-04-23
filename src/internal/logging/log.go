package logging

import (
	"fmt"
	"go.uber.org/zap"
)

var Instance *zap.SugaredLogger

func init() {
	config := zap.NewDevelopmentConfig()
	build, err := config.Build()
	if err != nil {
		panic(fmt.Errorf("zap logger setup failed: %w", err))
	}
	Instance = build.Sugar()
}

func Close() {
	Instance.Sync()
}
