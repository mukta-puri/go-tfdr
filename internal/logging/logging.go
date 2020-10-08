package logging

import (
	"github.com/sirupsen/logrus"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/config"
)

// InitLogger sets up logging level, and log formatting
func InitLogger() {
	c := config.GetConfig()
	ll, err := logrus.ParseLevel(c.LogLevel)
	if err != nil {
		ll = logrus.InfoLevel
	}
	logrus.SetLevel(ll)

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:  true,
		PadLevelText:   true,
		DisableQuote:   true,
		DisableSorting: true,
	})
}
