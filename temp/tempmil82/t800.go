package tempmil82

import (
	"context"
	"fmt"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
)

type T800 struct {
	p *comport.Port
}

func (x *T800) Close() error {
	if x.p != nil {
		return x.p.Close()
	}
	return nil
}

func (x *T800) SetComport(p *comport.Port) {
	x.p = p
}

func (x *T800) Start(log comm.Logger, ctx context.Context) error {
	_, err := getResponseStr(log, ctx, x.p, "01WRD,01,0101,0001")
	return err
}

func (x *T800) Stop(log comm.Logger, ctx context.Context) error {
	_, err := getResponseStr(log, ctx, x.p, "01WRD,01,0101,0004")
	return err
}

func (x *T800) Setup(log comm.Logger, ctx context.Context, value float64) error {
	v := int64(value * 10)
	s := fmt.Sprintf("01WRD,01,0102,%04X", v)
	_, err := getResponseStr(log, ctx, x.p, s)
	return err
}

func (x *T800) Read(log comm.Logger, ctx context.Context) (float64, error) {
	return getResponseStr(log, ctx, x.p, "01RRD,02,0001,0002")
}
