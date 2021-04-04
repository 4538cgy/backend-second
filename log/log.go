package log

import (
	"github.com/4538cgy/backend-second/config"
	"os"
	"time"

	"fmt"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"runtime"
)

var Trace = logrus.Trace
var Tracef = logrus.Tracef
var Debug = logrus.Debug
var Debugf = logrus.Debugf
var Info = logrus.Info
var Infof = logrus.Infof
var Warning = logrus.Warning
var Warningf = logrus.Warningf
var Error = logrus.Error
var Errorf = logrus.Errorf
var Fatal = logrus.Fatal
var Fatalf = logrus.Fatalf
var Panic = logrus.Panic
var Panicf = logrus.Panicf

func Init(cfg *config.Config) {
	if !cfg.Log.Enable {
		return
	}

	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
		ForceColors:     true,
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			return "", fmt.Sprintf(" %s:%d", filepath.Base(frame.File), frame.Line)
		},
	})

	switch cfg.Log.Level {
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "waring":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.SetOutput(&cfg.LogConfig)
	if cfg.Log.StdOut {
		logrus.SetOutput(os.Stdout)
	}

	return
}
