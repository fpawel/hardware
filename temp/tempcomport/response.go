package tempcomport

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/pkg/errors"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var Err = merry.New("ошибка термокамеры")

type ResponseReader struct {
	Wr io.ReadWriter
	C  comm.Config
	H  HandleResponseFunc
}

type HandleResponseFunc = func(request, response string)

func (rdr ResponseReader) getResponse(log comm.Logger, ctx context.Context, strRequest string) (float64, error) {
	strRequest = fmt.Sprintf("\x02%s\r\n", strRequest)
	var temperature float64
	response, err := comm.GetResponse(log, ctx, rdr.C, rdr.Wr, []byte(strRequest), func(_, response []byte) (string, error) {
		err := checkResponse(strRequest, response, &temperature)
		if rdr.H != nil {
			rdr.H(strRequest, string(response))
		}
		return string(response), err
	})
	if err != nil {
		err = merry.Appendf(err, "запрос=%q", strRequest).WithCause(Err)
		if len(response) > 0 {
			err = merry.Appendf(err, "ответ=%q", string(response))
		}
		return 0, err
	}
	return temperature, nil
}

var regexTemperature = regexp.MustCompile(`^01RRD,OK,([0-9a-fA-F]{4}),([0-9a-fA-F]{4})$`)

func checkResponse(strRequest string, response []byte, t *float64) error {
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

	r := strResponse[1 : len(strResponse)-2]

	if strings.HasPrefix(strRequest, "01WRD") && r != "01WRD,OK" {
		return merry.New("ответ на запрос 01WRD должен быть 01WRD,OK")
	}

	if strings.HasPrefix(strRequest, "01RRD") {
		if !strings.HasPrefix(r, "01RRD,OK") {
			return merry.New("не удалось считать температуру: ответ на запрос 01RRD должен начинаться со строки 01RRD,OK")
		}
		xs := regexTemperature.FindAllStringSubmatch(r, -1)
		if len(xs) == 0 {
			return merry.New("не правильный формат температуры")
		}
		if len(xs[1]) == 2 {
			return merry.New("не правильный формат температуры: ожидался код значения температуры и уставки")
		}
		n, err := strconv.ParseInt(xs[1][1], 16, 64)
		if err != nil {
			err = errors.Wrapf(err, "не правильный формат температуры: %q", xs[1][1])
			return merry.Wrap(err)
		}
		*t = float64(n) / 10
	}
	return nil
}
