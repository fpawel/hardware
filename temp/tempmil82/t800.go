package tempmil82

import (
	"context"
	"fmt"
	"github.com/fpawel/comm"
	"github.com/fpawel/hardware/temp"
)

type T800 struct {
	r ResponseReader
}

func NewT800(r ResponseReader) temp.TemperatureDevice {
	return T800{r: r}
}

func (x T800) Start(log comm.Logger, ctx context.Context) error {
	_, err := getResponse(log, ctx, x.r, "01WRD,01,0101,0001")
	return err
}

func (x T800) Stop(log comm.Logger, ctx context.Context) error {
	_, err := getResponse(log, ctx, x.r, "01WRD,01,0101,0004")
	return err
}

func (x T800) Setup(log comm.Logger, ctx context.Context, value float64) error {
	v := int64(value * 10)
	s := fmt.Sprintf("01WRD,01,0102,%04X", v)
	_, err := getResponse(log, ctx, x.r, s)
	return err
}

func (x T800) Read(log comm.Logger, ctx context.Context) (float64, error) {
	return getResponse(log, ctx, x.r, "01RRD,02,0001,0002")
}
