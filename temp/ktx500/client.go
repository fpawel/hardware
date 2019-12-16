package ktx500

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/gofins/fins"
	"github.com/fpawel/hardware/internal/pkg"
	"time"
)

type Client struct {
	c Config
}

func (x Client) Start(log comm.Logger, ctx context.Context) error {
	return x.write(log, ctx, "старт", func(fc *fins.Client) error {
		return fc.BitTwiddle(fins.MemoryAreaWRBit, 0, 0, true)
	})
}

func (x Client) Stop(log comm.Logger, ctx context.Context) error {
	return x.write(log, ctx, "стоп", func(fc *fins.Client) error {
		return fc.BitTwiddle(fins.MemoryAreaWRBit, 0, 0, false)
	})
}

func (x Client) Setup(log comm.Logger, ctx context.Context, temperature float64) error {
	return x.write(log, ctx, "стоп", func(fc *fins.Client) error {
		return finsWriteFloat(fc, 8, temperature)
	})
}

func (x Client) Read(log comm.Logger, ctx context.Context) (temperature float64, err error) {
	err = x.do(log, ctx, "запрос температуры", func(c *fins.Client) (string, error) {
		return fmt.Sprintf("%v", temperature), readTemperature(c, &temperature)
	})
	return
}

func (x Client) CoolingOn(log comm.Logger, ctx context.Context) error {
	return x.write(log, ctx, "включение охлаждения", func(c *fins.Client) error {
		return c.BitTwiddle(fins.MemoryAreaWRBit, 0, 10, true)
	})
}

func (x Client) CoolingOff(log comm.Logger, ctx context.Context) error {
	return x.write(log, ctx, "выключение охлаждения", func(c *fins.Client) error {
		return c.BitTwiddle(fins.MemoryAreaWRBit, 0, 10, false)
	})
}

func (x Client) write(log comm.Logger, ctx context.Context, what string, work func(*fins.Client) error) error {
	return x.do(log, ctx, what, func(fc *fins.Client) (string, error) {
		return "ok", work(fc)
	})
}

func (x Client) do(log comm.Logger, ctx context.Context, what string, work func(*fins.Client) (string, error)) error {

	log = pkg.LogPrependSuffixKeys(log, "действие", what)

	client, err := x.c.newClient()
	if err != nil {
		return wrapErr(err).Appendf("%s: установка соединения", what)
	}
	defer client.Close()

	for attempt := 0; attempt < x.c.MaxAttemptsRead; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var s string
		if s, err = work(client); err == nil {
			log.Info(s)
			return nil
		}
		log.PrintErr(merry.Appendf(err, "попытка %d из %d", attempt+1, x.c.MaxAttemptsRead))
		pause(ctx.Done(), time.Second)
	}
	return wrapErr(err)
}

func pause(chDone <-chan struct{}, d time.Duration) {
	timer := time.NewTimer(d)
	for {
		select {
		case <-timer.C:
			return
		case <-chDone:
			timer.Stop()
			return
		}
	}
}