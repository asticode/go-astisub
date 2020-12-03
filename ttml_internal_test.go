package astisub

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTTMLDuration(t *testing.T) {
	// Unmarshal hh:mm:ss.mmm format - clock time
	var d = &TTMLInDuration{}
	err := d.UnmarshalText([]byte("12:34:56.789"))
	assert.NoError(t, err)
	assert.Equal(t, 12*time.Hour+34*time.Minute+56*time.Second+789*time.Millisecond, d.duration())

	// Marshal
	b, err := TTMLOutDuration(d.duration()).MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, "12:34:56.789", string(b))

	// Unmarshal hh:mm:ss:fff format
	err = d.UnmarshalText([]byte("12:34:56:2"))
	assert.NoError(t, err)
	assert.Equal(t, 12*time.Hour+34*time.Minute+56*time.Second, d.duration())
	assert.Equal(t, 2, d.frames)

	// Duration
	d.framerate = 8
	assert.Equal(t, 12*time.Hour+34*time.Minute+56*time.Second+250*time.Millisecond, d.duration())

	// Unmarshal offset time
	err = d.UnmarshalText([]byte("123h"))
	assert.Equal(t, 123*time.Hour, d.duration())
	assert.NoError(t, err)

	err = d.UnmarshalText([]byte("123.4h"))
	assert.Equal(t, 123*time.Hour+4*time.Hour/10, d.duration())
	assert.NoError(t, err)

	err = d.UnmarshalText([]byte("123m"))
	assert.Equal(t, 123*time.Minute, d.duration())
	assert.NoError(t, err)

	err = d.UnmarshalText([]byte("123.4m"))
	assert.Equal(t, 123*time.Minute+4*time.Minute/10, d.duration())
	assert.NoError(t, err)

	err = d.UnmarshalText([]byte("123s"))
	assert.Equal(t, 123*time.Second, d.duration())
	assert.NoError(t, err)

	err = d.UnmarshalText([]byte("123.4s"))
	assert.Equal(t, 123*time.Second+4*time.Second/10, d.duration())
	assert.NoError(t, err)

	err = d.UnmarshalText([]byte("123ms"))
	assert.Equal(t, 123*time.Millisecond, d.duration())
	assert.NoError(t, err)

	err = d.UnmarshalText([]byte("123.4ms"))
	assert.Equal(t, 123*time.Millisecond+4*time.Millisecond/10, d.duration())
	assert.NoError(t, err)

	d.framerate = 25
	err = d.UnmarshalText([]byte("100f"))
	assert.Equal(t, 4*time.Second, d.duration())
	assert.NoError(t, err)

	// Tick rate duration
	d.tickrate = 4
	err = d.UnmarshalText([]byte("6t"))
	assert.Equal(t, time.Second+500*time.Millisecond, d.duration())
	assert.NoError(t, err)
}
