package astisub_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/asticode/go-astisub"
	"github.com/stretchr/testify/assert"
)

func TestSRT(t *testing.T) {
	// Open
	s, err := astisub.OpenFile("./testdata/example-in.srt")
	assert.NoError(t, err)
	assertSubtitleItems(t, s)

	// No subtitles to write
	w := &bytes.Buffer{}
	err = astisub.Subtitles{}.WriteToSRT(w)
	assert.EqualError(t, err, astisub.ErrNoSubtitlesToWrite.Error())

	// Write
	c, err := ioutil.ReadFile("./testdata/example-out.srt")
	assert.NoError(t, err)
	err = s.WriteToSRT(w)
	assert.NoError(t, err)
	assert.Equal(t, string(c), w.String())
}

func TestSRTMissingSequence(t *testing.T) {
	// Open
	s, err := astisub.OpenFile("./testdata/missing-sequence-in.srt")
	assert.NoError(t, err)
	assertSubtitleItems(t, s)

	// No subtitles to write
	w := &bytes.Buffer{}
	err = astisub.Subtitles{}.WriteToSRT(w)
	assert.EqualError(t, err, astisub.ErrNoSubtitlesToWrite.Error())

	// Write
	c, err := ioutil.ReadFile("./testdata/example-out.srt")
	assert.NoError(t, err)
	err = s.WriteToSRT(w)
	assert.NoError(t, err)
	assert.Equal(t, string(c), w.String())
}

func TestNonUTF8SRT(t *testing.T) {
	_, err := astisub.OpenFile("./testdata/example-in-non-utf8.srt")
	assert.Error(t, err)
}

func TestSRTStyled(t *testing.T) {
	// Open
	s, err := astisub.OpenFile("./testdata/example-in-styled.srt")
	assert.NoError(t, err)

	// assert the items are properly parsed
	assert.Len(t, s.Items, 6)
	assert.Equal(t, 17*time.Second+985*time.Millisecond, s.Items[0].StartAt)
	assert.Equal(t, 20*time.Second+521*time.Millisecond, s.Items[0].EndAt)
	assert.Equal(t, "[instrumental music]", s.Items[0].Lines[0].String())
	assert.Equal(t, 47*time.Second+115*time.Millisecond, s.Items[1].StartAt)
	assert.Equal(t, 48*time.Second+282*time.Millisecond, s.Items[1].EndAt)
	assert.Equal(t, "[ticks]", s.Items[1].Lines[0].String())
	assert.Equal(t, 58*time.Second+192*time.Millisecond, s.Items[2].StartAt)
	assert.Equal(t, 59*time.Second+727*time.Millisecond, s.Items[2].EndAt)
	assert.Equal(t, "[instrumental music]", s.Items[2].Lines[0].String())
	assert.Equal(t, 1*time.Minute+1*time.Second+662*time.Millisecond, s.Items[3].StartAt)
	assert.Equal(t, 1*time.Minute+3*time.Second+63*time.Millisecond, s.Items[3].EndAt)
	assert.Equal(t, "[dog barking]", s.Items[3].Lines[0].String())
	assert.Equal(t, 1*time.Minute+26*time.Second+787*time.Millisecond, s.Items[4].StartAt)
	assert.Equal(t, 1*time.Minute+29*time.Second+523*time.Millisecond, s.Items[4].EndAt)
	assert.Equal(t, "[beeping]", s.Items[4].Lines[0].String())
	assert.Equal(t, 1*time.Minute+29*time.Second+590*time.Millisecond, s.Items[5].StartAt)
	assert.Equal(t, 1*time.Minute+31*time.Second+992*time.Millisecond, s.Items[5].EndAt)
	assert.Equal(t, "[automated]", s.Items[5].Lines[0].String())
	assert.Equal(t, "'The time is 7:35.'", s.Items[5].Lines[1].String())

	// assert the styles of the items
	assert.Len(t, s.Items, 6)
	assert.Equal(t, "#00ff00", *s.Items[0].Lines[0].Items[0].InlineStyle.SRTColor)
	assert.Zero(t, s.Items[0].Lines[0].Items[0].InlineStyle.SRTPosition)
	assert.True(t, s.Items[0].Lines[0].Items[0].InlineStyle.SRTBold)
	assert.False(t, s.Items[0].Lines[0].Items[0].InlineStyle.SRTItalics)
	assert.False(t, s.Items[0].Lines[0].Items[0].InlineStyle.SRTUnderline)
	assert.Equal(t, "#ff00ff", *s.Items[1].Lines[0].Items[0].InlineStyle.SRTColor)
	assert.Zero(t, s.Items[1].Lines[0].Items[0].InlineStyle.SRTPosition)
	assert.False(t, s.Items[1].Lines[0].Items[0].InlineStyle.SRTBold)
	assert.False(t, s.Items[1].Lines[0].Items[0].InlineStyle.SRTItalics)
	assert.False(t, s.Items[1].Lines[0].Items[0].InlineStyle.SRTUnderline)
	assert.Equal(t, "#00ff00", *s.Items[2].Lines[0].Items[0].InlineStyle.SRTColor)
	assert.Zero(t, s.Items[2].Lines[0].Items[0].InlineStyle.SRTPosition)
	assert.False(t, s.Items[2].Lines[0].Items[0].InlineStyle.SRTBold)
	assert.False(t, s.Items[2].Lines[0].Items[0].InlineStyle.SRTItalics)
	assert.False(t, s.Items[2].Lines[0].Items[0].InlineStyle.SRTUnderline)
	assert.Nil(t, s.Items[3].Lines[0].Items[0].InlineStyle.SRTColor)
	assert.Zero(t, s.Items[3].Lines[0].Items[0].InlineStyle.SRTPosition)
	assert.True(t, s.Items[3].Lines[0].Items[0].InlineStyle.SRTBold)
	assert.False(t, s.Items[3].Lines[0].Items[0].InlineStyle.SRTItalics)
	assert.True(t, s.Items[3].Lines[0].Items[0].InlineStyle.SRTUnderline)
	assert.Nil(t, s.Items[4].Lines[0].Items[0].InlineStyle)
	assert.Nil(t, s.Items[5].Lines[0].Items[0].InlineStyle)
	assert.Nil(t, s.Items[5].Lines[1].Items[0].InlineStyle.SRTColor)
	assert.Zero(t, s.Items[5].Lines[1].Items[0].InlineStyle.SRTPosition)
	assert.False(t, s.Items[5].Lines[1].Items[0].InlineStyle.SRTBold)
	assert.True(t, s.Items[5].Lines[1].Items[0].InlineStyle.SRTItalics)
	assert.False(t, s.Items[5].Lines[1].Items[0].InlineStyle.SRTUnderline)

	// Write to srt
	w := &bytes.Buffer{}
	c, err := os.ReadFile("./testdata/example-out-styled.srt")
	assert.NoError(t, err)
	err = s.WriteToSRT(w)
	assert.NoError(t, err)
	assert.Equal(t, string(c), w.String())

	// Write to WebVTT
	w = &bytes.Buffer{}
	c, err = os.ReadFile("./testdata/example-out-styled.vtt")
	assert.NoError(t, err)
	err = s.WriteToWebVTT(w)
	assert.NoError(t, err)
	assert.Equal(t, string(c), w.String())
}
