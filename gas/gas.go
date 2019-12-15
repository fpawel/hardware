package gas

import (
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
)

type DevType int

const (
	DevTypeMil82 DevType = iota
	DevTypeLab73CO
)

func New(t DevType) Switcher{
	switch t {
	case DevTypeMil82:
		return gasMil82{}
	case DevTypeLab73CO:
		return gasLab73CO{}
	default:
		panic(t)
	}
}

type Switcher interface {
	Switch(log comm.Logger, rdr modbus.ResponseReader, addr modbus.Addr, n byte, ) error
}

type gasMil82 struct {}

func (_ gasMil82) Switch(log comm.Logger, rdr modbus.ResponseReader, addr modbus.Addr, n byte, ) error{
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

type gasLab73CO struct {}

func (_ gasLab73CO) Switch(log comm.Logger, rdr modbus.ResponseReader, addr modbus.Addr, n byte, ) error{
	req := modbus.Request{
		Addr:     5,
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