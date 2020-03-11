package tempcomport

import (
	"context"
	"fmt"
	"github.com/fpawel/comm"
	"github.com/fpawel/hardware/temp"
)

type T2500 comm.T

var _ temp.TemperatureDevice = T2500{}

func (x T2500) Start(log comm.Logger, ctx context.Context) error {
	return getResponse(log, ctx, "старт", comm.T(x), "01WRD,01,0102,0001", nil)
}

func (x T2500) Stop(log comm.Logger, ctx context.Context) error {
	return getResponse(log, ctx, "стоп", comm.T(x), "01WRD,01,0102,0004", nil)
}

func (x T2500) Setup(log comm.Logger, ctx context.Context, value float64) error {
	v := int64(value * 10)
	s := fmt.Sprintf("01WRD,01,0104,%04X", v)
	return getResponse(log, ctx, fmt.Sprintf("уставка %v⁰C", value), comm.T(x), s, nil)
}

func (x T2500) Read(log comm.Logger, ctx context.Context) (temperature float64, err error) {
	err = getResponse(log, ctx, "запрос температуры", comm.T(x), "01RRD,02,0001,0002", &temperature)
	return
}
