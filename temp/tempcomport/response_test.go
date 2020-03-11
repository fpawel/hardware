package tempcomport

import (
	"context"
	"github.com/fpawel/comm"
	"github.com/powerman/structlog"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetResponse(t *testing.T) {

	testReadTemperature(t, []byte{
		0x02, 0x30, 0x31, 0x52, 0x52, 0x44, 0x2C, 0x4F, 0x4B, 0x2C,
		0x46, 0x45, 0x38, 0x34, 0x2C, 0x30, 0x30,
		0x43, 0x38, 0x0D, 0x0A,
	}, -38)
	testReadTemperature(t, []byte{
		0x02, 0x30, 0x31, 0x52, 0x52, 0x44, 0x2C, 0x4F, 0x4B, 0x2C,
		0x30, 0x33, 0x32, 0x30, 0x2C, 0x30, 0x33,
		0x32, 0x30, 0x0D, 0x0A,
	}, 80)
	testReadTemperature(t, []byte{
		0x02, 0x30, 0x31, 0x52, 0x52, 0x44, 0x2C, 0x4F, 0x4B, 0x2C,
		0x46, 0x46, 0x34, 0x46, 0x2C, 0x30, 0x30,
		0x43, 0x38, 0x0D, 0x0A,
	}, -17.7)

	for n := float64(-999); n <= 999; n++ {
		v := n / 10
		x, err := parseTemperature(formatTemperature(v))
		assert.NoError(t, err)
		assert.Equal(t, v, x)
	}
}

func testReadTemperature(t *testing.T, b []byte, value float64) {
	var temperature float64
	cm := comm.New(rw{b}, comm.Config{
		TimeoutGetResponse: 1,
	})
	assert.NoError(t, getResponse(structlog.New(), context.Background(), "", cm, "01RRD,02,0001,0002", &temperature))
	assert.Equal(t, value, temperature)
	v, err := parseTemperature(formatTemperature(temperature))
	assert.NoError(t, err)
	assert.Equal(t, value, v)
}

type rw struct {
	d []byte
}

func (x rw) Read(p []byte) (int, error) {

	if len(p) >= len(x.d) {
		copy(p, x.d)
	}
	return len(x.d), nil
}

func (rw) Write(p []byte) (int, error) {
	return len(p), nil
}

func init() {
	comm.SetEnableLog(false)
}
