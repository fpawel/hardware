package tempcomport

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/hardware/internal/pkg"
	"github.com/pkg/errors"
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

var regexTemperature = regexp.MustCompile(`^01RRD,OK,([0-9a-fA-F]{4}),([0-9a-fA-F]{4})$`)

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

	r := strResponse[1 : len(strResponse)-2]

	if !strings.HasPrefix(r, "01RRD,OK") {
		return merry.New("не удалось считать температуру: ответ на запрос температуры 01RRD должен начинаться со строки 01RRD,OK")
	}

	xs := regexTemperature.FindAllStringSubmatch(r, -1)
	if len(xs) == 0 {
		return merry.New("не правильный формат температуры")
	}
	if len(xs[0]) != 3 {
		return merry.New("не правильный формат температуры: ожидался код значения температуры и уставки")
	}

	str := xs[0][1]

	n, err := strconv.ParseInt(str, 16, 17)
	if err != nil {
		err = errors.Wrapf(err, "не правильный формат температуры: %q", str)
		return merry.Wrap(err)
	}
	*temperature = float64(int16(n)) / 10

	return nil
}
