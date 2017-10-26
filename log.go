package wechat

import "go.uber.org/zap"

var Logger *zap.Logger
var Sugar *zap.SugaredLogger

func InitLogger(config ...zap.Config) {
	var conf zap.Config
	if len(config) > 1 {
		conf = config[0]
	} else {
		conf = zap.NewDevelopmentConfig()
		conf.DisableStacktrace = true
	}
	var err error
	Logger, err = conf.Build()
	if err != nil {
		panic(err)
	}
	Sugar = Logger.Sugar()
}
