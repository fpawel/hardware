package tempcomport

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/hardware/internal/pkg"
	"regexp"
	"strconv"
	"strings"
)

var Err = merry.New("ошибка термокамеры")

type HandleResponseFunc = func(request, response string)

func getResponse(log comm.Logger, ctx context.Context, what string, cm comm.T, strRequest string, temperature *float64) error {
	log = pkg.LogPrependSuffixKeys(log,
		"request_temperature_device", strRequest,
		"temperature_device_command", fmt.Sprintf("`%s`", what),
	)
	strRawRequest := fmt.Sprintf("\x02%s\r\n", strRequest)

	wrapErr := func(response []byte, err error) error {
		err = merry.Prependf(err, "термокамера: %s: request_temperature_device=%q", what, strRawRequest)
		if len(response) > 0 {
			err = merry.Prependf(err, "response_temperature_device=%q", string(response))
		}
		return merry.WithCause(err, Err)
	}
	response, err := cm.GetResponse(log, ctx, []byte(strRawRequest))
	if err != nil {
		return wrapErr(response, err)
	}
	if err := checkResponse(response, temperature); err != nil {
		return wrapErr(response, err)
	}
	return nil
}

func formatTemperature(v float64) string {
	s := fmt.Sprintf("%04X", uint32(int16(v*10)))
	s = s[len(s)-4:]
	return s
}

func parseTemperature(s string) (float64, error) {
	n, err := strconv.ParseInt(s, 16, 17)
	if err != nil {
		return 0, err
	}
	return float64(int16(n)) / 10, nil
}

var regexTemperature = regexp.MustCompile(`^\x0201RRD,OK,([\da-fA-F]{4}),([\da-fA-F]{4})\r\n$`)

func parseTemperatureResponse(s string) (float64, error) {
	xs := regexTemperature.FindAllStringSubmatch(s, -1)
	if len(xs) == 0 {
		return 0, merry.New("не правильный формат значения температуры")
	}
	if len(xs[0]) != 3 {
		return 0, merry.New("не правильный формат значения емпературы: ожидался код значения температуры и уставки")
	}
	return parseTemperature(xs[0][1])
}

func checkResponse(response []byte, temperature *float64) error {

	if len(response) == 4 {
		return merry.New("нет ответа от термокамеры")
	}

	if len(response) < 4 {
		return merry.New("длина ответа менее 4")
	}
	if response[0] != 2 {
		return merry.New("первый байт ответа не 2")
	}

	strResponse := string(response)

	if !strings.HasSuffix(strResponse, "\r\n") {
		return merry.New("ответ должен оканчиваться байтами 0D 0A")
	}

	if temperature == nil {
		return nil
	}

	var err error
	*temperature, err = parseTemperatureResponse(strResponse)
	return err

}
