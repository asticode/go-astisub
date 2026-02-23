package astisub

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestColor(t *testing.T) {
	c, err := newColorFromSSAString("305419896", 10)
	assert.NoError(t, err)
	assert.Equal(t, Color{Alpha: 0x12, Blue: 0x34, Green: 0x56, Red: 0x78}, *c)
	c, err = newColorFromSSAString("12345678", 16)
	assert.NoError(t, err)
	assert.Equal(t, Color{Alpha: 0x12, Blue: 0x34, Green: 0x56, Red: 0x78}, *c)
	assert.Equal(t, "#785634", c.HTMLString())
	assert.Equal(t, "12345678", c.SSAString())
}

func TestParseDuration(t *testing.T) {
	_, err := parseDuration("12:34:56,1234", ",", 3)
	assert.EqualError(t, err, "astisub: Invalid number of millisecond digits detected in 12:34:56,1234")
	_, err = parseDuration("12,123", ",", 3)
	assert.EqualError(t, err, "astisub: No hours, minutes or seconds detected in 12,123")
	d, err := parseDuration("12:34,123", ",", 3)
	assert.NoError(t, err)
	assert.Equal(t, 12*time.Minute+34*time.Second+123*time.Millisecond, d)
	d, err = parseDuration("12:34:56,123", ",", 3)
	assert.NoError(t, err)
	assert.Equal(t, 12*time.Hour+34*time.Minute+56*time.Second+123*time.Millisecond, d)
	d, err = parseDuration("12:34:56,1", ",", 3)
	assert.NoError(t, err)
	assert.Equal(t, 12*time.Hour+34*time.Minute+56*time.Second+100*time.Millisecond, d)
	d, err = parseDuration("12:34:56.123", ".", 3)
	assert.NoError(t, err)
	assert.Equal(t, 12*time.Hour+34*time.Minute+56*time.Second+123*time.Millisecond, d)
	d, err = parseDuration("1:23:45.67", ".", 2)
	assert.NoError(t, err)
	assert.Equal(t, time.Hour+23*time.Minute+45*time.Second+67*time.Millisecond, d)
	d, err = parseDuration("1:23:45:67", ":", 2)
	assert.NoError(t, err)
	assert.Equal(t, time.Hour+23*time.Minute+45*time.Second+67*time.Millisecond, d)
}

func TestFormatDuration(t *testing.T) {
	s := formatDuration(time.Second, ",", 3)
	assert.Equal(t, "00:00:01,000", s)
	s = formatDuration(time.Second, ",", 2)
	assert.Equal(t, "00:00:01,00", s)
	s = formatDuration(time.Millisecond, ",", 3)
	assert.Equal(t, "00:00:00,001", s)
	s = formatDuration(10*time.Millisecond, ".", 3)
	assert.Equal(t, "00:00:00.010", s)
	s = formatDuration(100*time.Millisecond, ",", 3)
	assert.Equal(t, "00:00:00,100", s)
	s = formatDuration(time.Second+234*time.Millisecond, ",", 3)
	assert.Equal(t, "00:00:01,234", s)
	s = formatDuration(12*time.Second+345*time.Millisecond, ",", 3)
	assert.Equal(t, "00:00:12,345", s)
	s = formatDuration(2*time.Minute+3*time.Second+456*time.Millisecond, ",", 3)
	assert.Equal(t, "00:02:03,456", s)
	s = formatDuration(20*time.Minute+34*time.Second+567*time.Millisecond, ",", 3)
	assert.Equal(t, "00:20:34,567", s)
	s = formatDuration(3*time.Hour+25*time.Minute+45*time.Second+678*time.Millisecond, ",", 3)
	assert.Equal(t, "03:25:45,678", s)
	s = formatDuration(34*time.Hour+17*time.Minute+36*time.Second+789*time.Millisecond, ",", 3)
	assert.Equal(t, "34:17:36,789", s)
	s = formatDuration(12*time.Hour+34*time.Minute+56*time.Second+999*time.Millisecond, ",", 2)
	assert.Equal(t, "12:34:56,99", s)
}

func TestPropagateSTLAttributes(t *testing.T) {
	// Test JustificationRight propagates to both WebVTTAlign and TTMLTextAlign
	t.Run("JustificationRight", func(t *testing.T) {
		sa := &StyleAttributes{
			STLJustification: &JustificationRight,
		}
		sa.propagateSTLAttributes()
		assert.Equal(t, "right", sa.WebVTTAlign)
		assert.NotNil(t, sa.TTMLTextAlign)
		assert.Equal(t, "right", *sa.TTMLTextAlign)
	})

	// Test JustificationLeft propagates to both WebVTTAlign and TTMLTextAlign
	t.Run("JustificationLeft", func(t *testing.T) {
		sa := &StyleAttributes{
			STLJustification: &JustificationLeft,
		}
		sa.propagateSTLAttributes()
		assert.Equal(t, "left", sa.WebVTTAlign)
		assert.NotNil(t, sa.TTMLTextAlign)
		assert.Equal(t, "left", *sa.TTMLTextAlign)
	})

	// Test JustificationCentered doesn't set TTMLTextAlign
	t.Run("JustificationCentered", func(t *testing.T) {
		sa := &StyleAttributes{
			STLJustification: &JustificationCentered,
		}
		sa.propagateSTLAttributes()
		assert.Empty(t, sa.WebVTTAlign)
		assert.Nil(t, sa.TTMLTextAlign)
	})

	// Test nil justification doesn't set anything
	t.Run("NoJustification", func(t *testing.T) {
		sa := &StyleAttributes{}
		sa.propagateSTLAttributes()
		assert.Empty(t, sa.WebVTTAlign)
		assert.Nil(t, sa.TTMLTextAlign)
	})

	// Test STLPosition to WebVTTLine conversion (in-vision)
	t.Run("STLPositionInVision", func(t *testing.T) {
		sa := &StyleAttributes{
			STLPosition: &STLPosition{
				VerticalPosition: 50,
				MaxRows:          99,
			},
		}
		sa.propagateSTLAttributes()
		assert.Equal(t, "50%", sa.WebVTTLine)
	})

	// Test STLPosition to WebVTTLine conversion (teletext)
	t.Run("STLPositionTeletext", func(t *testing.T) {
		sa := &StyleAttributes{
			STLPosition: &STLPosition{
				VerticalPosition: 22,
				MaxRows:          23,
			},
		}
		sa.propagateSTLAttributes()
		// (22-1)*100/23 = 91
		assert.Equal(t, "91%", sa.WebVTTLine)
	})

	// Test STLPosition at row 0 for teletext (edge case)
	t.Run("STLPositionTeletextRow0", func(t *testing.T) {
		sa := &StyleAttributes{
			STLPosition: &STLPosition{
				VerticalPosition: 0,
				MaxRows:          23,
			},
		}
		sa.propagateSTLAttributes()
		// When VerticalPosition is 0, we don't subtract 1
		assert.Equal(t, "0%", sa.WebVTTLine)
	})

	// Test combined justification and position
	t.Run("JustificationAndPosition", func(t *testing.T) {
		sa := &StyleAttributes{
			STLJustification: &JustificationRight,
			STLPosition: &STLPosition{
				VerticalPosition: 10,
				MaxRows:          23,
			},
		}
		sa.propagateSTLAttributes()
		assert.Equal(t, "right", sa.WebVTTAlign)
		assert.NotNil(t, sa.TTMLTextAlign)
		assert.Equal(t, "right", *sa.TTMLTextAlign)
		assert.Equal(t, "39%", sa.WebVTTLine) // (10-1)*100/23 = 39
	})
}
