package tempmil82

import (
	"context"
	"fmt"
	"github.com/fpawel/comm"
	"github.com/fpawel/hardware/temp"
)

type T2500 struct {
	r comm.ResponseReader
}

func NewT2500(r comm.ResponseReader) temp.TemperatureDevice {
	return T2500{r: r}
}

func (x T2500) Start(log comm.Logger, ctx context.Context) error {
	_, err := getResponse(log, ctx, x.r, "01WRD,01,0102,0001")
	return err
}

func (x T2500) Stop(log comm.Logger, ctx context.Context) error {
	_, err := getResponse(log, ctx, x.r, "01WRD,01,0102,0004")
	return err
}

func (x T2500) Setup(log comm.Logger, ctx context.Context, value float64) error {
	v := int64(value * 10)
	s := fmt.Sprintf("01WRD,01,0104,%04X", v)
	_, err := getResponse(log, ctx, x.r, s)
	return err
}

func (x T2500) Read(log comm.Logger, ctx context.Context) (float64, error) {
	return getResponse(log, ctx, x.r, "01RRD,02,0001,0002")
}
