package astisub

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSTLDuration(t *testing.T) {
	// Default
	d, err := parseDurationSTL("12345678", 100)
	assert.NoError(t, err)
	assert.Equal(t, 12*time.Hour+34*time.Minute+56*time.Second+780*time.Millisecond, d)
	s := formatDurationSTL(d, 100)
	assert.Equal(t, "12345678", s)

	// Bytes
	b := formatDurationSTLBytes(d, 100)
	assert.Equal(t, []byte{0xc, 0x22, 0x38, 0x4e}, b)
	d2 := parseDurationSTLBytes([]byte{0xc, 0x22, 0x38, 0x4e}, 100)
	assert.Equal(t, d, d2)
}
