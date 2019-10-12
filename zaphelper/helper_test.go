package zaphelper

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	logPath = "helper.log"
)

func TestBuildRotateLogger(t *testing.T) {
	_ = os.Remove(logPath)
	_, err := os.Stat(logPath)
	if !os.IsNotExist(err) {
		t.FailNow()
	}

	conf := zap.NewProductionConfig()
	conf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	conf.EncoderConfig.TimeKey = "timestamp"
	conf.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	hc := RotatingFileConfig{
		Filename: logPath,
	}

	logger := BuildRotateLogger(conf, hc)
	logger.Info("hello world")
	logger.Debug("debug info")
	logger.Warn("warning")
	_ = logger.Sync()

	_, err = os.Stat(logPath)
	if os.IsNotExist(err) {
		t.FailNow()
	}

	file, _ := os.Open(logPath)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var count int
	for scanner.Scan() {
		count++
		line := make(map[string]interface{})
		err = json.Unmarshal(scanner.Bytes(), &line)
		if err != nil {
			t.FailNow()
		}
	}

	if count != 3 {
		t.FailNow()
	}
}
