package astisub

import (
	"testing"

	"time"

	"github.com/asticode/go-astitools/ptr"
	"github.com/stretchr/testify/assert"
)

func TestTeletextPESDataType(t *testing.T) {
	m := make(map[int]string)
	for i := 0; i < 255; i++ {
		t := teletextPESDataType(uint8(i))
		if t != teletextPESDataTypeUnknown {
			m[i] = t
		}
	}
	assert.Equal(t, map[int]string{19: teletextPESDataTypeEBU, 20: teletextPESDataTypeEBU, 21: teletextPESDataTypeEBU, 26: teletextPESDataTypeEBU, 28: teletextPESDataTypeEBU, 17: teletextPESDataTypeEBU, 27: teletextPESDataTypeEBU, 31: teletextPESDataTypeEBU, 16: teletextPESDataTypeEBU, 18: teletextPESDataTypeEBU, 23: teletextPESDataTypeEBU, 29: teletextPESDataTypeEBU, 22: teletextPESDataTypeEBU, 24: teletextPESDataTypeEBU, 25: teletextPESDataTypeEBU, 30: teletextPESDataTypeEBU}, m)
}

func TestTeletextPageParse(t *testing.T) {
	p := newTeletextPage(0, time.Unix(10, 0))
	p.end = time.Unix(15, 0)
	p.rows = []int{2, 1}
	p.data = map[uint8][]byte{
		1: append([]byte{0xb}, []byte("test1")...),
		2: append([]byte{0xb}, []byte("test2")...),
	}
	s := Subtitles{}
	d := newTeletextCharacterDecoder()
	d.updateCharset(astiptr.UInt8(0), false)
	p.parse(&s, d, time.Unix(5, 0))
	assert.Equal(t, []*Item{{
		EndAt: 10 * time.Second,
		Lines: []Line{
			{Items: []LineItem{{InlineStyle: &StyleAttributes{TeletextSpacesAfter: astiptr.Int(0), TeletextSpacesBefore: astiptr.Int(0)}, Text: "test1"}}},
			{Items: []LineItem{{InlineStyle: &StyleAttributes{TeletextSpacesAfter: astiptr.Int(0), TeletextSpacesBefore: astiptr.Int(0)}, Text: "test2"}}},
		},
		StartAt: 5 * time.Second,
	}}, s.Items)
}

func TestTeletextPageParseRow(t *testing.T) {
	b := []byte("start")
	b = append(b, 0x0, 0xb)
	b = append(b, []byte("black")...)
	b = append(b, 0x1)
	b = append(b, []byte("red")...)
	b = append(b, 0x2)
	b = append(b, []byte("green")...)
	b = append(b, 0x3)
	b = append(b, []byte("yellow")...)
	b = append(b, 0x4)
	b = append(b, []byte("blue")...)
	b = append(b, 0x5)
	b = append(b, []byte("magenta")...)
	b = append(b, 0x6)
	b = append(b, []byte("cyan")...)
	b = append(b, 0x7)
	b = append(b, []byte("white")...)
	b = append(b, 0xd)
	b = append(b, []byte("double height")...)
	b = append(b, 0xe)
	b = append(b, []byte("double width")...)
	b = append(b, 0xf)
	b = append(b, []byte("double size")...)
	b = append(b, 0xc)
	b = append(b, []byte("reset")...)
	b = append(b, 0xa)
	b = append(b, []byte("end")...)
	i := Item{}
	d := newTeletextCharacterDecoder()
	d.updateCharset(astiptr.UInt8(0), false)
	p := teletextPage{}
	p.parseRow(&i, d, b)
	assert.Equal(t, 1, len(i.Lines))
	assert.Equal(t, []LineItem{
		{Text: "black", InlineStyle: &StyleAttributes{
			TeletextColor:        ColorBlack,
			TeletextSpacesAfter:  astiptr.Int(0),
			TeletextSpacesBefore: astiptr.Int(0),
		}},
		{Text: "red", InlineStyle: &StyleAttributes{
			TeletextColor:        ColorRed,
			TeletextSpacesAfter:  astiptr.Int(0),
			TeletextSpacesBefore: astiptr.Int(0),
		}},
		{Text: "green", InlineStyle: &StyleAttributes{
			TeletextColor:        ColorGreen,
			TeletextSpacesAfter:  astiptr.Int(0),
			TeletextSpacesBefore: astiptr.Int(0),
		}},
		{Text: "yellow", InlineStyle: &StyleAttributes{
			TeletextColor:        ColorYellow,
			TeletextSpacesAfter:  astiptr.Int(0),
			TeletextSpacesBefore: astiptr.Int(0),
		}},
		{Text: "blue", InlineStyle: &StyleAttributes{
			TeletextColor:        ColorBlue,
			TeletextSpacesAfter:  astiptr.Int(0),
			TeletextSpacesBefore: astiptr.Int(0),
		}},
		{Text: "magenta", InlineStyle: &StyleAttributes{
			TeletextColor:        ColorMagenta,
			TeletextSpacesAfter:  astiptr.Int(0),
			TeletextSpacesBefore: astiptr.Int(0),
		}},
		{Text: "cyan", InlineStyle: &StyleAttributes{
			TeletextColor:        ColorCyan,
			TeletextSpacesAfter:  astiptr.Int(0),
			TeletextSpacesBefore: astiptr.Int(0),
		}},
		{Text: "white", InlineStyle: &StyleAttributes{
			TeletextColor:        ColorWhite,
			TeletextSpacesAfter:  astiptr.Int(0),
			TeletextSpacesBefore: astiptr.Int(0),
		}},
		{Text: "double height", InlineStyle: &StyleAttributes{
			TeletextColor:        ColorWhite,
			TeletextDoubleHeight: astiptr.Bool(true),
			TeletextSpacesAfter:  astiptr.Int(0),
			TeletextSpacesBefore: astiptr.Int(0),
		}},
		{Text: "double width", InlineStyle: &StyleAttributes{
			TeletextColor:        ColorWhite,
			TeletextDoubleHeight: astiptr.Bool(true),
			TeletextDoubleWidth:  astiptr.Bool(true),
			TeletextSpacesAfter:  astiptr.Int(0),
			TeletextSpacesBefore: astiptr.Int(0),
		}},
		{Text: "double size", InlineStyle: &StyleAttributes{
			TeletextColor:        ColorWhite,
			TeletextDoubleHeight: astiptr.Bool(true),
			TeletextDoubleWidth:  astiptr.Bool(true),
			TeletextDoubleSize:   astiptr.Bool(true),
			TeletextSpacesAfter:  astiptr.Int(0),
			TeletextSpacesBefore: astiptr.Int(0),
		}},
		{Text: "reset", InlineStyle: &StyleAttributes{
			TeletextColor:        ColorWhite,
			TeletextDoubleHeight: astiptr.Bool(false),
			TeletextDoubleWidth:  astiptr.Bool(false),
			TeletextDoubleSize:   astiptr.Bool(false),
			TeletextSpacesAfter:  astiptr.Int(0),
			TeletextSpacesBefore: astiptr.Int(0),
		}},
	}, i.Lines[0].Items)
}

func TestTeletextPageAppendLineItem(t *testing.T) {
	// Init
	l := Line{}
	p := &teletextPage{}

	// Empty
	p.appendLineItem(&l, LineItem{})
	assert.Equal(t, 0, len(l.Items))

	// Not empty
	p.appendLineItem(&l, LineItem{Text: " test  "})
	assert.Equal(t, "test", l.Items[0].Text)
	assert.Equal(t, StyleAttributes{
		TeletextSpacesAfter:  astiptr.Int(2),
		TeletextSpacesBefore: astiptr.Int(1),
	}, *l.Items[0].InlineStyle)
}
