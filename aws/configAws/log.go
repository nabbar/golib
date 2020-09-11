package configAws

import (
	"reflect"

	"github.com/nabbar/golib/logger"
)

type awsLogger struct {
	logLevel logger.Level
}

func (l awsLogger) Log(args ...interface{}) {
	pattern := ""

	for i := 0; i < len(args); i++ {
		//nolint #exhaustive
		switch reflect.TypeOf(args[i]).Kind() {
		case reflect.String:
			pattern += "%s"
		default:
			pattern += "%v"
		}
	}

	l.logLevel.Logf("AWS Log : "+pattern, args...)
}

func LevelPanic() logger.Level {
	return logger.PanicLevel
}

func LevelFatal() logger.Level {
	return logger.FatalLevel
}

func LevelError() logger.Level {
	return logger.ErrorLevel
}

func LevelWarn() logger.Level {
	return logger.WarnLevel
}

func LevelInfo() logger.Level {
	return logger.InfoLevel
}

func LevelDebug() logger.Level {
	return logger.DebugLevel
}

func LevelNoLog() logger.Level {
	return logger.NilLevel
}
