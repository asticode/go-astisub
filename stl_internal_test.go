package astisub

import (
	"testing"
	"time"

	"github.com/asticode/go-astikit"
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

func TestSTLCharacterHandler(t *testing.T) {
	h, err := newSTLCharacterHandler(stlCharacterCodeTableNumberLatin)
	assert.NoError(t, err)
	o := h.decode(0x1f)
	assert.Equal(t, []byte(nil), o)
	o = h.decode(0x65)
	assert.Equal(t, []byte("e"), o)
	o = h.decode(0xc1)
	assert.Equal(t, []byte(nil), o)
	o = h.decode(0x65)
	assert.Equal(t, []byte("è"), o)
}

func TestSTLCharacterHandlerUmlaut(t *testing.T) {
	h, err := newSTLCharacterHandler(stlCharacterCodeTableNumberLatin)
	assert.NoError(t, err)

	o := h.decode(0xc8)
	assert.Equal(t, []byte(nil), o)
	o = h.decode(0x61)
	assert.Equal(t, []byte("ä"), o)

	o = h.decode(0xc8)
	assert.Equal(t, []byte(nil), o)
	o = h.decode(0x41)
	assert.Equal(t, []byte("Ä"), o)

	o = h.decode(0xc8)
	assert.Equal(t, []byte(nil), o)
	o = h.decode(0x6f)
	assert.Equal(t, []byte("ö"), o)

	o = h.decode(0xc8)
	assert.Equal(t, []byte(nil), o)
	o = h.decode(0x4f)
	assert.Equal(t, []byte("Ö"), o)

	o = h.decode(0xc8)
	assert.Equal(t, []byte(nil), o)
	o = h.decode(0x75)
	assert.Equal(t, []byte("ü"), o)

	o = h.decode(0xc8)
	assert.Equal(t, []byte(nil), o)
	o = h.decode(0x55)
	assert.Equal(t, []byte("Ü"), o)

	o = h.decode(0xc8)
	assert.Equal(t, []byte(nil), o)
	o = h.decode(0x65)
	assert.Equal(t, []byte("ë"), o)

	o = h.decode(0xc8)
	assert.Equal(t, []byte(nil), o)
	o = h.decode(0x45)
	assert.Equal(t, []byte("Ë"), o)

	o = h.decode(0xc8)
	assert.Equal(t, []byte(nil), o)
	o = h.decode(0x69)
	assert.Equal(t, []byte("ï"), o)

	o = h.decode(0xc8)
	assert.Equal(t, []byte(nil), o)
	o = h.decode(0x49)
	assert.Equal(t, []byte("Ï"), o)
}

func TestSTLStyler(t *testing.T) {
	// Parse spacing attributes
	s := newSTLStyler()
	s.parseSpacingAttribute(0x80)
	assert.Equal(t, stlStyler{italics: astikit.BoolPtr(true)}, *s)
	s.parseSpacingAttribute(0x81)
	assert.Equal(t, stlStyler{italics: astikit.BoolPtr(false)}, *s)
	s = newSTLStyler()
	s.parseSpacingAttribute(0x82)
	assert.Equal(t, stlStyler{underline: astikit.BoolPtr(true)}, *s)
	s.parseSpacingAttribute(0x83)
	assert.Equal(t, stlStyler{underline: astikit.BoolPtr(false)}, *s)
	s = newSTLStyler()
	s.parseSpacingAttribute(0x84)
	assert.Equal(t, stlStyler{boxing: astikit.BoolPtr(true)}, *s)
	s.parseSpacingAttribute(0x85)
	assert.Equal(t, stlStyler{boxing: astikit.BoolPtr(false)}, *s)

	// Has been set
	s = newSTLStyler()
	assert.False(t, s.hasBeenSet())
	s.boxing = astikit.BoolPtr(true)
	assert.True(t, s.hasBeenSet())
	s = newSTLStyler()
	s.italics = astikit.BoolPtr(true)
	assert.True(t, s.hasBeenSet())
	s = newSTLStyler()
	s.underline = astikit.BoolPtr(true)
	assert.True(t, s.hasBeenSet())

	// Has changed
	s = newSTLStyler()
	sa := &StyleAttributes{}
	assert.False(t, s.hasChanged(sa))
	s.boxing = astikit.BoolPtr(true)
	assert.True(t, s.hasChanged(sa))
	sa.STLBoxing = s.boxing
	assert.False(t, s.hasChanged(sa))
	s.italics = astikit.BoolPtr(true)
	assert.True(t, s.hasChanged(sa))
	sa.STLItalics = s.italics
	assert.False(t, s.hasChanged(sa))
	s.underline = astikit.BoolPtr(true)
	assert.True(t, s.hasChanged(sa))
	sa.STLUnderline = s.underline
	assert.False(t, s.hasChanged(sa))

	// Update
	s = newSTLStyler()
	sa = &StyleAttributes{}
	s.update(sa)
	assert.Equal(t, StyleAttributes{}, *sa)
	s.boxing = astikit.BoolPtr(true)
	s.update(sa)
	assert.Equal(t, StyleAttributes{STLBoxing: s.boxing}, *sa)
	s.italics = astikit.BoolPtr(true)
	s.update(sa)
	assert.Equal(t, StyleAttributes{STLBoxing: s.boxing, STLItalics: s.italics}, *sa)
	s.underline = astikit.BoolPtr(true)
	s.update(sa)
	assert.Equal(t, StyleAttributes{STLBoxing: s.boxing, STLItalics: s.italics, STLUnderline: s.underline}, *sa)
}
