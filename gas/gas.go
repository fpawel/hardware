package gas

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/hardware/internal/pkg"
)

type DevType int

const (
	Mil82 DevType = iota
	Lab73CO
)

func Switch(devType DevType, log comm.Logger, rdr modbus.ResponseReader, addr modbus.Addr, n byte) error {

	log = pkg.LogPrependSuffixKeys(log,
		"тип_газового_блока", devType.String(),
		"адрес_газового_блока", addr,
		"клапан", n)
	wrapErr := func(err error) error {
		if err == nil {
			return nil
		}
		return merry.Appendf(err, "тип газового блока: %s, адрес газового блока: %d, клапан: %d",
			devType, addr, n)
	}

	d, err := devType.newSwitcher()
	if err != nil {
		return wrapErr(err)
	}
	return wrapErr(d.Switch(log, rdr, addr, n))
}

func (t DevType) String() string {
	switch t {
	case Mil82:
		return "МИЛ82"
	case Lab73CO:
		return "Лаб73СО"
	default:
		return fmt.Sprintf("%d", t)
	}
}

func (t DevType) newSwitcher() (switcher, error) {
	switch t {
	case Mil82:
		return gasMil82{}, nil
	case Lab73CO:
		return gasLab73CO{}, nil
	default:
		return nil, merry.Errorf("не правильный тип пневмолока: %d", t)
	}
}

type switcher interface {
	Switch(log comm.Logger, rdr modbus.ResponseReader, addr modbus.Addr, n byte) error
}

type gasMil82 struct{}

func (_ gasMil82) Switch(log comm.Logger, rdr modbus.ResponseReader, addr modbus.Addr, n byte) error {
	req := modbus.Request{
		Addr:     addr,
		ProtoCmd: 0x10,
		Data: []byte{
			0, 0x10, 0, 1, 2, 0, n,
		},
	}
	_, err := rdr.GetResponse(req.Bytes(), log, nil)
	return err
}

type gasLab73CO struct{}

func (_ gasLab73CO) Switch(log comm.Logger, rdr modbus.ResponseReader, addr modbus.Addr, n byte) error {
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
		return merry.Errorf("не правильный код клапана: %d", n)
	}
	if _, err := rdr.GetResponse(req.Bytes(), log, nil); err != nil {
		return merry.Append(err, "переключение клапана")
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

	if _, err := rdr.GetResponse(req.Bytes(), log, nil); err != nil {
		return merry.Append(err, "установка расхода")
	}

	return nil
}
