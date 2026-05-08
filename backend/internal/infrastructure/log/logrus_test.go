package log

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestInitJSONAndLevel(t *testing.T) {
	Init("debug", "json")
	if L().Level != logrus.DebugLevel {
		t.Errorf("level=%v", L().Level)
	}
	if _, ok := L().Formatter.(*logrus.JSONFormatter); !ok {
		t.Errorf("expected JSON formatter, got %T", L().Formatter)
	}
}

func TestInitTextDefault(t *testing.T) {
	Init("garbage", "text")
	if L().Level != logrus.InfoLevel {
		t.Errorf("level=%v", L().Level)
	}
	if _, ok := L().Formatter.(*logrus.TextFormatter); !ok {
		t.Errorf("expected Text formatter, got %T", L().Formatter)
	}
}

func TestParseLevels(t *testing.T) {
	tests := map[string]logrus.Level{
		"trace":   logrus.TraceLevel,
		"debug":   logrus.DebugLevel,
		"warn":    logrus.WarnLevel,
		"warning": logrus.WarnLevel,
		"error":   logrus.ErrorLevel,
		"fatal":   logrus.FatalLevel,
		"":        logrus.InfoLevel,
		"weird":   logrus.InfoLevel,
	}
	for in, want := range tests {
		if got := parseLevel(in); got != want {
			t.Errorf("parseLevel(%q)=%v want %v", in, got, want)
		}
	}
}
