package features

import (
	syslog "../syslog"
	xenstoreclient "../xenstoreclient"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type FeatureIPSettingClient interface {
	Run() error
}

type FeatureIPSetting struct {
	Client  xenstoreclient.XenStoreClient
	Enabled bool
	Debug   bool
	logger  *log.Logger
}

const (
	advertiseKey = "control/feature-static-ip-setting"
	controlKey   = "xenserver/device/vif"
	token        = "FeatureIPSetting"
)

const (
	LoggerName string = "FeatureIPSetting"
)

func NewFeatureIPSetting(Client xenstoreclient.XenStoreClient, Enabled bool, Debug bool) (FeatureIPSettingClient, error) {
	var loggerWriter io.Writer = os.Stderr
	var topic string = LoggerName
	if w, err := syslog.NewSyslogWriter(topic); err == nil {
		loggerWriter = w
		topic = ""
	} else {
		fmt.Fprintf(os.Stderr, "NewSyslogWriter(%s) error: %s, use stderr logging\n", topic, err)
		topic = LoggerName + ": "
	}
	logger := log.New(loggerWriter, topic, 0)

	return &FeatureIPSetting{
		Client:  Client,
		Enabled: Enabled,
		Debug:   Debug,
		logger:  logger,
	}, nil
}

func (f *FeatureIPSetting) Enable() {
	if f.Enabled {
		f.Client.Write(advertiseKey, "1")
	} else {
		f.Client.Write(advertiseKey, "0")
	}
	return
}

func (f *FeatureIPSetting) Run() error {
	err := f.Client.Watch(controlKey, token)
	if err != nil {
		f.logger.Printf("Watch(\"%#v\") error: %#v\n", controlKey, err)
		return err
	}

	f.logger.Printf("Start watch on %#v\n", controlKey)
	go func() {
		ticker := time.Tick(4 * time.Second)
		for {
			f.Enable()
			if _, ok := f.Client.WatchEvent(controlKey); ok {
				f.logger.Printf("featureIPSetting get event")
			}
			select {
			case <-ticker:
				continue
			}

		}
	}()
	return nil
}
