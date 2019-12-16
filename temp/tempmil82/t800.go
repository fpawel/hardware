package tempmil82

import (
	"context"
	"fmt"
	"github.com/fpawel/comm"
)

type T800 struct {
	ResponseReader
}

func (x T800) Start(log comm.Logger, ctx context.Context) error {
	_, err := getResponseMil82(log, ctx, x.ResponseReader, "01WRD,01,0101,0001")
	return err
}

func (x T800) Stop(log comm.Logger, ctx context.Context) error {
	_, err := getResponseMil82(log, ctx, x.ResponseReader, "01WRD,01,0101,0004")
	return err
}

func (x T800) Setup(log comm.Logger, ctx context.Context, value float64) error {
	v := int64(value * 10)
	s := fmt.Sprintf("01WRD,01,0102,%04X", v)
	_, err := getResponseMil82(log, ctx, x.ResponseReader, s)
	return err
}

func (x T800) Read(log comm.Logger, ctx context.Context) (float64, error) {
	return getResponseMil82(log, ctx, x.ResponseReader, "01RRD,02,0001,0002")
}
