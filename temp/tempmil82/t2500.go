package tempmil82

import (
	"fmt"
	"github.com/fpawel/hardware/temp"
)

type T2500 struct{ r ResponseReader }

func (x T2500) Start(log temp.Logger, ctx temp.Ctx) error {
	_, err := getResponseMil82(log, ctx, x.r, "01WRD,01,0102,0001")
	return err
}

func (x T2500) Stop(log temp.Logger, ctx temp.Ctx) error {
	_, err := getResponseMil82(log, ctx, x.r, "01WRD,01,0102,0004")
	return err
}

func (x T2500) Setup(log temp.Logger, ctx temp.Ctx, value float64) error {
	v := int64(value * 10)
	s := fmt.Sprintf("01WRD,01,0104,%04X", v)
	_, err := getResponseMil82(log, ctx, x.r, s)
	return err
}

func (x T2500) Read(log temp.Logger, ctx temp.Ctx) (float64, error) {
	return getResponseMil82(log, ctx, x.r, "01RRD,02,0001,0002")
}
