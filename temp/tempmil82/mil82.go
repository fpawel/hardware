package tempmil82

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/pkg/errors"
	"regexp"
	"strconv"
	"strings"
)

var Err = merry.New("ошибка термокамеры")

func getResponse(log comm.Logger, ctx context.Context, rdr comm.ResponseReader, s string) (float64, error) {
	s = fmt.Sprintf("\x02%s\r\n", s)
	b, err := rdr.GetResponse(log, ctx, []byte(s))

	if err != nil {
		return 0, newErr("нет связи", s, nil)
	}
	if len(b) < 4 {
		return 0, newErr("несоответствие протоколу: длина ответа менее 4", s, b)
	}
	if b[0] != 2 {
		return 0, newErr("несоответствие протоколу: первый байт ответа не 2", s, b)
	}

	r := string(b)

	if !strings.HasSuffix(r, "\r\n") {
		return 0, newErr("несоответствие протоколу: ответ должен оканчиваться байтами 0D 0A", s, b)
	}

	r = r[1 : len(r)-2]

	if strings.HasPrefix(s, "01WRD") && r != "01WRD,OK" {
		return 0, newErr("несоответствие протоколу: ответ на запрос 01WRD должен быть 01WRD,OK", s, b)
	}

	if strings.HasPrefix(s, "01RRD") {
		if !strings.HasPrefix(r, "01RRD,OK") {
			return 0, newErr("несоответствие протоколу: не удалось считать температуру: ответ на запрос 01RRD должен начинаться со строки 01RRD,OK", s, b)
		}
		xs := regexTemperature.FindAllStringSubmatch(r, -1)
		if len(xs) == 0 {
			return 0, newErr("не правильный формат температуры", s, b)
		}
		if len(xs[1]) == 2 {
			return 0, newErr("не правильный формат температуры: ожидался код значения температуры и уставки", s, b)
		}
		n, err := strconv.ParseInt(xs[1][1], 16, 64)
		if err != nil {
			err = errors.Wrapf(err, "не правильный формат температуры: %q", xs[1][1])
			return 0, wrapErr(err, s, b)
		}

		return float64(n) / 10, nil
	}
	return 0, nil
}

var regexTemperature = regexp.MustCompile(`^01RRD,OK,([0-9a-fA-F]{4}),([0-9a-fA-F]{4})$`)

func newErr(err string, strReq string, b []byte) error {
	return wrapErr(merry.New(err), strReq, b)
}

func wrapErr(err error, strReq string, b []byte) error {
	return merry.Appendf(err, "%v: запрос %q: [% X], ответ %q: [% X]",
		err, strReq, []byte(strReq), string(b), b).WithCause(Err)
}
