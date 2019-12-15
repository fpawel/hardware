package ktx500

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/internal/api"
	"github.com/fpawel/elco/internal/cfg"
	"github.com/fpawel/gofins/fins"
	"github.com/powerman/structlog"
	"math"
	"sync"
	"time"
)

var (
	Err = merry.New("КТХ-500")
)

type Info struct {
	Temperature, Destination float64
	TemperatureOn, CoolOn    bool
}

type FinsNetwork struct {
	MaxAttemptsRead int           `yaml:"max_attempts_read" comment:"число попыток получения ответа"`
	TimeoutMS       uint          `yaml:"timeout_ms" comment:"таймаут считывания, мс"`
	Pause           time.Duration `yaml:"pause" comment:"пауза опроса, с"`
	Server          FinsSettings  `yaml:"server" comment:"параметры ссервера omron fins"`
	Client          FinsSettings  `yaml:"client" comment:"параметры клиента omron fins"`
}

type FinsSettings struct {
	IP       string `yaml:"ip" comment:"tcp адрес"`
	Port     int    `yaml:"port" comment:"tcp порт"`
	Network  byte   `yaml:"network" comment:"fins network"`
	Node     byte   `yaml:"node" comment:"fins node"`
	FinsUnit byte   `yaml:"unit" comment:"fins unit"`
}

func ReadTemperature() (temperature float64, err error) {
	err = write("запрос температуры", func(c *fins.Client) error {
		return readTemperature(c, &temperature)
	})
	return
}

func WriteDestination(value float64) error {
	return write(fmt.Sprintf("запись уставки %v⁰C", value), func(c *fins.Client) error {
		return finsWriteFloat(c, 8, value)
	})
}

func WriteOnOff(value bool) error {
	s := "включение уставки"
	if !value {
		s = "выключение уставки"
	}
	return write(s, func(c *fins.Client) error {
		return c.BitTwiddle(fins.MemoryAreaWRBit, 0, 0, value)
	})
}

func WriteCoolOnOff(value bool) error {
	s := "включение компрессора"
	if !value {
		s = "выключение компрессора"
	}
	return write(s, func(c *fins.Client) error {
		return c.BitTwiddle(fins.MemoryAreaWRBit, 0, 10, value)
	})
}

func GetLast() (Info, error) {
	last.Mutex.Lock()
	defer last.Mutex.Unlock()
	if last.error != nil {
		return Info{}, last.error
	}
	return last.Ktx500Info, nil
}

func SetupTemperature(destinationTemperature float64) error {

	err := func() error {
		// запись уставки
		if err := WriteDestination(destinationTemperature); err != nil {
			return err
		}
		// включение термокамеры
		if err := WriteOnOff(true); err != nil {
			return err
		}

		// установка компрессора
		if err := WriteCoolOnOff(destinationTemperature < 50); err != nil {
			return err
		}
		return nil
	}()

	return merry.Appendf(err, "установка %v⁰C", destinationTemperature)
}

func wrapErr(err error) merry.Error {
	if merry.Is(err, Err) {
		return merry.Wrap(err)
	}
	return merry.WithCause(err, Err)
}

func read(client *fins.Client, config FinsNetwork) error {

	var err error
	for attempt := 0; attempt < config.MaxAttemptsRead; attempt++ {
		if err = f(client); err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return err
}

func write(client *fins.Client, config FinsNetwork, what string) error {
	var err error
	for attempt := 0; attempt < config.MaxAttemptsRead; attempt++ {
		if err = f(client); err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		log.PrintErr(merry.Append(err, what))
		return err
	}
	log.Info(what + ": ok")
	return nil
}

func withClient(config FinsNetwork, work func(*fins.Client, FinsNetwork) error) error {
	client, err := config.NewFinsClient()
	if err != nil {
		return wrapErr(err).Append("установка соединения")
	}
	defer client.Close()
	if err = work(client, config); err != nil {
		return wrapErr(err)
	}
	return nil
}

func readTemperature(c *fins.Client, temperature *float64) (err error) {
	*temperature, err = finsReadFloat(c, 2)
	if err != nil {
		return wrapErr(err).Append("запрос температуры")
	}
	return
}

func readInfo(x *Info) error {
	return read(func(c *fins.Client) error {
		var (
			coolOn, temperatureOn []bool
			temperature           float64
		)

		err := readTemperature(c, &temperature)
		if err != nil {
			return err
		}

		destination, err := finsReadFloat(c, 8)
		if err != nil {
			return wrapErr(err).Append("запрос уставки")
		}

		temperatureOn, err = c.ReadBits(fins.MemoryAreaWRBit, 0, 0, 1)
		if err != nil {
			return wrapErr(err).Append("запрос состояния термокамеры")
		}

		coolOn, err = c.ReadBits(fins.MemoryAreaWRBit, 0, 10, 1)
		if err != nil {
			return wrapErr(err).Append("запрос состояния компрессора")
		}

		*x = Info{
			Temperature:   math.Round(temperature*100.) / 100.,
			Destination:   destination,
			TemperatureOn: temperatureOn[0],
			CoolOn:        coolOn[0],
		}
		return nil
	})
}

func eqNfo(x, y Info) bool {
	if x == y {
		return true
	}
	a, b := x, y
	a.Temperature, b.Temperature = 0, 0
	return a == b && math.Abs(x.Temperature-y.Temperature) < 0.5
}

func finsReadFloat(finsClient *fins.Client, addr int) (float64, error) {
	xs, err := finsClient.ReadBytes(fins.MemoryAreaDMWord, uint16(addr), 2)
	if err != nil {
		return 0, err
	}
	buf := bytes.NewBuffer([]byte{xs[1], xs[0], xs[3], xs[2]})
	var v float32
	if err := binary.Read(buf, binary.LittleEndian, &v); err != nil {
		return 0, err
	}
	return float64(v), nil
}

func finsWriteFloat(finsClient *fins.Client, addr int, value float64) error {

	buf := bytes.NewBuffer(nil)
	_ = binary.Write(buf, binary.LittleEndian, math.Float32bits(float32(value)))
	b := buf.Bytes()
	words := []uint16{
		binary.LittleEndian.Uint16([]byte{b[0], b[1]}),
		binary.LittleEndian.Uint16([]byte{b[2], b[3]}),
	}
	return finsClient.WriteWords(fins.MemoryAreaDMWord, uint16(addr), words)
}

var (
	last struct {
		sync.Mutex
		Info
		error
	}
	log      = structlog.New()
	muClient sync.Mutex
)
