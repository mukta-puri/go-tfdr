package logging

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/config"
)

func TestInitLogger(t *testing.T) {
	os.Setenv("TF_STATE_COPY_LOG_LEVEL", "panic")
	config.InitConfig("./no-file")
	InitLogger()
	assert.Equal(t, logrus.PanicLevel, logrus.GetLevel())
}

func TestDefault(t *testing.T) {
	os.Setenv("TF_STATE_COPY_LOG_LEVEL", "not-a-real-log-level")
	config.InitConfig("./no-file")
	InitLogger()
	assert.Equal(t, logrus.InfoLevel, logrus.GetLevel())
}
