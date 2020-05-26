package gas

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/hardware/internal/pkg"
)

type DevType string

const (
	Mil82   DevType = "МИЛ82"
	Lab73CO DevType = "Лаб73СО"
)

func Switch(log comm.Logger, ctx context.Context, devType DevType, cm comm.T, addr modbus.Addr, n byte) error {
	log = pkg.LogPrependSuffixKeys(log,
		"пневмоблок_тип", devType,
		"пневмоблок_адрес", addr,
		"пневмоблок_переключение", n)
	wrapErr := func(err error) error {
		if err == nil {
			return nil
		}
		return merry.Prependf(err,
			"пневмоблок_переключение=%d пневмоблок_адрес=%d пневмоблок_тип=%s",
			n, addr, devType)
	}
	d, err := devType.newSwitcher()
	if err != nil {
		return wrapErr(err)
	}
	return wrapErr(d.Switch(log, ctx, cm, addr, n))
}

func (t DevType) newSwitcher() (switcher, error) {
	switch t {
	case Mil82:
		return gasMil82{}, nil
	case Lab73CO:
		return gasLab73CO{}, nil
	default:
		return nil, merry.Errorf("не правильный тип пневмоблока %q", t)
	}
}

type switcher interface {
	Switch(log comm.Logger, ctx context.Context, cm comm.T, addr modbus.Addr, n byte) error
}

type gasMil82 struct{}

func (_ gasMil82) Switch(log comm.Logger, ctx context.Context, cm comm.T, addr modbus.Addr, n byte) error {
	req := modbus.Request{
		Addr:     addr,
		ProtoCmd: 0x10,
		Data: []byte{
			0, 0x10, 0, 1, 2, 0, n,
		},
	}
	_, err := req.GetResponse(log, ctx, cm)
	return err
}

type gasLab73CO struct{}

func (_ gasLab73CO) Switch(log comm.Logger, ctx context.Context, cm comm.T, addr modbus.Addr, n byte) error {
	req := modbus.Request{
		Addr:     addr,
		ProtoCmd: 0x10,
		Data: []byte{
			0, 0x32, 0, 1, 2, 0, 0,
		},
	}
	switch n {
	case 0:
		req.Data[6] = 0
	case 1:
		req.Data[6] = 1
	case 2:
		req.Data[6] = 2
	case 3:
		req.Data[6] = 4
	default:
		return merry.Errorf("не правильный код переключения пневмоблока: %d", n)
	}
	if _, err := req.GetResponse(pkg.LogPrependSuffixKeys(log, "пневмоблок_переключение", n), ctx, cm); err != nil {
		return merry.Appendf(err, "переключение %d", n)
	}

	req = modbus.Request{
		Addr:     1,
		ProtoCmd: 6,
		Data: []byte{
			0, 4, 0, 0,
		},
	}
	if n > 0 {
		req.Data[2] = 0x14
		req.Data[3] = 0xD5
	}

	if _, err := req.GetResponse(pkg.LogPrependSuffixKeys(log, "пневмоблок_установка_расхода", "D514"), ctx, cm); err != nil {
		return merry.Prepend(err, "установка расхода D514")
	}

	return nil
}
