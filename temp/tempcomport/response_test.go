package tempcomport

import (
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
	assert.Equal(t, formatTemperature(-40), "FE70")
	assert.Equal(t, formatTemperature(-40.5), "FE6B")
	assert.Equal(t, formatTemperature(-60.5), "FDA3")
	assert.Equal(t, formatTemperature(80.7), "0327")

}

func testReadTemperature(t *testing.T, b []byte, value float64) {
	temperature, err := parseTemperatureResponse(string(b))
	assert.NoError(t, err)
	assert.Equal(t, value, temperature)
	v, err := parseTemperature(formatTemperature(temperature))
	assert.NoError(t, err)
	assert.Equal(t, value, v)
}
