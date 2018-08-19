package p1

import (
	"context"

	"github.com/TIBCOSoftware/flogo-lib/core/trigger"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/cisco/senml"
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
		return log.Infof("no settings found for trigger '%s'", t.config.Id)
	}

	// Make sure the publishKey item exists
	if _, ok := t.config.Settings["serial_port"]; !ok {
		return log.Infof("no serial port found for trigger '%s' in settings", t.config.Id)
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
			msg := createSenML(r)
			logger.Debugf("%v", msg)
			for _, handler := range t.handlers {
				handler.Handle(context.Background(), map[string]interface{}{
					"msg":     msg,
					"KWh":     convertToFloat64(&r.Electricity.KWh),
					"KWhLow":  convertToFloat64(&r.Electricity.KWhLow),
					"W":       convertToFloat64(&r.Electricity.W),
					"GasUsed": convertToFloat64(&r.Gas.LastRecord.Value),
				})
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

func convertToFloat64(f32 *float32) *float64 {
	tt := float64(*f32)
	return &tt
}

func createSenML(t *dsmrp1.Telegram) string {
	s := senml.SenML{
		Records: []senml.SenMLRecord{
			senml.SenMLRecord{
				BaseName:    "Mill/P1/EnergyUsage/",
				BaseUnit:    "",
				BaseVersion: 5,
			},
			senml.SenMLRecord{Name: "W", Unit: "W", Value: convertToFloat64(&t.Electricity.W)},
			senml.SenMLRecord{Name: "KWh", Unit: "KWh", Value: convertToFloat64(&t.Electricity.KWh)},
			senml.SenMLRecord{Name: "KWhLow", Unit: "KWh", Value: convertToFloat64(&t.Electricity.KWhLow)},
			senml.SenMLRecord{Name: "GasUsed", Unit: "m3", Value: convertToFloat64(&t.Gas.LastRecord.Value)},
		},
	}

	n := senml.Normalize(s)

	dataOut, err := senml.Encode(n, senml.JSON, senml.OutputOptions{PrettyPrint: false})
	if err != nil {
		logger.Debugf(err.Error())
	}
	return string(dataOut)
}
