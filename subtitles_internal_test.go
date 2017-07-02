package astisub

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDuration(t *testing.T) {
	d, err := parseDuration("12:34:56,1234", ",")
	assert.EqualError(t, err, "Invalid number of millisecond digits detected in 12:34:56,1234")
	d, err = parseDuration("12,123", ",")
	assert.EqualError(t, err, "No hours, minutes or seconds detected in 12,123")
	d, err = parseDuration("12:34,123", ",")
	assert.NoError(t, err)
	assert.Equal(t, 12*time.Minute+34*time.Second+123*time.Millisecond, d)
	d, err = parseDuration("12:34:56,123", ",")
	assert.NoError(t, err)
	assert.Equal(t, 12*time.Hour+34*time.Minute+56*time.Second+123*time.Millisecond, d)
	d, err = parseDuration("12:34:56,1", ",")
	assert.NoError(t, err)
	assert.Equal(t, 12*time.Hour+34*time.Minute+56*time.Second+100*time.Millisecond, d)
	d, err = parseDuration("12:34:56.123", ".")
	assert.NoError(t, err)
	assert.Equal(t, 12*time.Hour+34*time.Minute+56*time.Second+123*time.Millisecond, d)
}

func TestFormatDuration(t *testing.T) {
	s := formatDuration(time.Second, ",")
	assert.Equal(t, "00:00:01,000", s)
	s = formatDuration(time.Millisecond, ",")
	assert.Equal(t, "00:00:00,001", s)
	s = formatDuration(10*time.Millisecond, ".")
	assert.Equal(t, "00:00:00.010", s)
	s = formatDuration(100*time.Millisecond, ",")
	assert.Equal(t, "00:00:00,100", s)
	s = formatDuration(time.Second+234*time.Millisecond, ",")
	assert.Equal(t, "00:00:01,234", s)
	s = formatDuration(12*time.Second+345*time.Millisecond, ",")
	assert.Equal(t, "00:00:12,345", s)
	s = formatDuration(2*time.Minute+3*time.Second+456*time.Millisecond, ",")
	assert.Equal(t, "00:02:03,456", s)
	s = formatDuration(20*time.Minute+34*time.Second+567*time.Millisecond, ",")
	assert.Equal(t, "00:20:34,567", s)
	s = formatDuration(3*time.Hour+25*time.Minute+45*time.Second+678*time.Millisecond, ",")
	assert.Equal(t, "03:25:45,678", s)
	s = formatDuration(34*time.Hour+17*time.Minute+36*time.Second+789*time.Millisecond, ",")
	assert.Equal(t, "34:17:36,789", s)
}
