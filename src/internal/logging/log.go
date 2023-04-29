package logging

import (
	"fmt"
	"go.uber.org/zap"
)

var Instance *zap.SugaredLogger

func init() {
	buildLogger("")
}

func buildLogger(output string) {
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zap.FatalLevel)
	if output != "" {
		config.OutputPaths = []string{output}
	}
	build, err := config.Build()
	if err != nil {
		panic(fmt.Errorf("zap logger setup failed: %w", err))
	}
	Instance = build.Sugar()
}

func Close() {
	Instance.Sync()
}
