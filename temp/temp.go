package temp

import (
	"context"
	"github.com/fpawel/comm"
)

type TemperatureDevice interface {
	Start(Logger, Ctx) error
	Stop(Logger, Ctx) error
	Setup(Logger, Ctx, float64) error
	Read(Logger, Ctx) (float64, error)
}

type Cooler interface {
	CoolingOn(Logger, Ctx) error
	CoolingOff(Logger, Ctx) error
}

type Logger = comm.Logger
type Ctx = context.Context
