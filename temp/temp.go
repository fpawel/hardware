package temp

import (
	"context"
	"github.com/fpawel/comm"
)

type TemperatureDevice interface {
	Start(comm.Logger, context.Context) error
	Stop(comm.Logger, context.Context) error
	Setup(comm.Logger, context.Context, float64) error
	Read(comm.Logger, context.Context) (float64, error)
}

type Cooler interface {
	CoolingOn(comm.Logger, context.Context) error
	CoolingOff(comm.Logger, context.Context) error
}
