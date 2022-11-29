package logger

import "go.uber.org/zap"

var logger *zap.SugaredLogger

func init() {
	unsugared, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	defer unsugared.Sync()
	logger = unsugared.Sugar()
}

func Debugw(msg string, keysAndValues ...interface{}) {
	logger.Debugw(msg, keysAndValues...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	logger.Infow(msg, keysAndValues...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	logger.Errorw(msg, keysAndValues...)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	logger.Fatalw(msg, keysAndValues...)
}
