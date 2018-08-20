package p1

import (
	"context"
	"fmt"

	"github.com/TIBCOSoftware/flogo-lib/core/trigger"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	dsmrp1 "github.com/rubenn/go-dsmrp1"
)

// log is the default package logger
var log = logger.GetLogger("trigger-p1")

// MyTriggerFactory My Trigger factory
type MyTriggerFactory struct {
	metadata *trigger.Metadata
}

// NewFactory create a new Trigger factory
func NewFactory(md *trigger.Metadata) trigger.Factory {
	return &MyTriggerFactory{metadata: md}
}

// New Creates a new trigger instance for a given id
func (t *MyTriggerFactory) New(config *trigger.Config) trigger.Trigger {
	return &MyTrigger{metadata: t.metadata, config: config}
}

// MyTrigger is a stub for your Trigger implementation
type MyTrigger struct {
	metadata *trigger.Metadata
	config   *trigger.Config
	handlers []*trigger.Handler
}

// Initialize implements trigger.Init.Initialize
func (t *MyTrigger) Initialize(ctx trigger.InitContext) error {
	if t.config.Settings == nil {
		return fmt.Errorf("no settings found for trigger '%s'", t.config.Id)
	}

	// Make sure the serial_port item exists
	if _, ok := t.config.Settings["serial_port"]; !ok {
		return fmt.Errorf("no serial port found for trigger '%s' in settings", t.config.Id)
	}

	t.handlers = ctx.GetHandlers()

	return nil
}

// Metadata implements trigger.Trigger.Metadata
func (t *MyTrigger) Metadata() *trigger.Metadata {
	return t.metadata
}

// Start implements trigger.Trigger.Start
func (t *MyTrigger) Start() error {
	// start the trigger
	m, err := dsmrp1.NewMeter("/dev/ttyUSB0")
	if err != nil {
		logger.Debugf("Failed to create meter: %v", err)
	}

	go func() {
		for r := range m.C {
			for _, handler := range t.handlers {
				trgData := make(map[string]interface{})
				trgData["KWh"] = r.Electricity.KWh
				trgData["KWhLow"] = r.Electricity.KWhLow
				trgData["W"] = r.Electricity.W
				trgData["GasUsed"] = r.Gas.LastRecord.Value

				results, err := handler.Handle(context.Background(), trgData)
				if err != nil {
					log.Error("Error starting action: ", err.Error())
				}

				log.Debugf("Ran Handler: [%s]", handler)
				log.Debugf("Results: [%v]", results)
			}
		}
	}()

	return nil
}

// Stop implements trigger.Trigger.Start
func (t *MyTrigger) Stop() error {
	// stop the trigger
	return nil
}
