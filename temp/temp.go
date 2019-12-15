package temp

import (
	"context"
	"github.com/fpawel/comm"
)

type Interface interface {
	Start(Logger, Ctx) error
	Stop(Logger, Ctx) error
	Setup(Logger, Ctx, float64) error
	Read(Logger, Ctx) (float64, error)
}

type Logger = comm.Logger
type Ctx = context.Context
