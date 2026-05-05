package config

import (
	"fmt"
	"sync"

	"github.com/newrelic/go-agent/v3/newrelic"
)

var initNewRelic = &sync.Once{}
var newRelicApp *newrelic.Application = nil

func NewRelicApp() *newrelic.Application {
	initNewRelic.Do(func() {
		var err error
		newRelicApp, err = newrelic.NewApplication(
			newrelic.ConfigEnabled(Cfg.NewRelicEnabled),
			newrelic.ConfigAppName(Cfg.NewRelicAppName),
			newrelic.ConfigLicense(Cfg.NewRelicLicenseKey),
			newrelic.ConfigDistributedTracerEnabled(true),
		)
		if err != nil {
			_ = fmt.Errorf("error setting up NewRelic: %v", err)
		}
	})
	return newRelicApp
}
