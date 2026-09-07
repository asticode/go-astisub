package astisub_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/asticode/go-astisub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Len(t, s.Items, 10)
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
	assert.Equal(t, "Test with multi line italics", s.Items[6].Lines[0].String())
	assert.Equal(t, "Terminated on the next line", s.Items[6].Lines[1].String())
	assert.Equal(t, "Unterminated styles", s.Items[7].Lines[0].String())
	assert.Equal(t, "Do no fall to the next item", s.Items[8].Lines[0].String())
	assert.Equal(t, "x", s.Items[9].Lines[0].Items[0].Text)
	assert.Equal(t, "^3 * ", s.Items[9].Lines[0].Items[1].Text)
	assert.Equal(t, "x", s.Items[9].Lines[0].Items[2].Text)
	assert.Equal(t, " = 100", s.Items[9].Lines[0].Items[3].Text)

	// assert the styles of the items
	assert.Equal(t, astisub.ColorLime, s.Items[0].Lines[0].Items[0].InlineStyle.SRTColor)
	assert.True(t, s.Items[0].Lines[0].Items[0].InlineStyle.SRTBold)
	assert.False(t, s.Items[0].Lines[0].Items[0].InlineStyle.SRTItalics)
	assert.False(t, s.Items[0].Lines[0].Items[0].InlineStyle.SRTUnderline)
	assert.Equal(t, astisub.ColorMagenta, s.Items[1].Lines[0].Items[0].InlineStyle.SRTColor)
	assert.False(t, s.Items[1].Lines[0].Items[0].InlineStyle.SRTBold)
	assert.False(t, s.Items[1].Lines[0].Items[0].InlineStyle.SRTItalics)
	assert.False(t, s.Items[1].Lines[0].Items[0].InlineStyle.SRTUnderline)
	assert.Equal(t, astisub.ColorLime, s.Items[2].Lines[0].Items[0].InlineStyle.SRTColor)
	assert.False(t, s.Items[2].Lines[0].Items[0].InlineStyle.SRTBold)
	assert.False(t, s.Items[2].Lines[0].Items[0].InlineStyle.SRTItalics)
	assert.False(t, s.Items[2].Lines[0].Items[0].InlineStyle.SRTUnderline)
	assert.Nil(t, s.Items[3].Lines[0].Items[0].InlineStyle.SRTColor)
	assert.True(t, s.Items[3].Lines[0].Items[0].InlineStyle.SRTBold)
	assert.False(t, s.Items[3].Lines[0].Items[0].InlineStyle.SRTItalics)
	assert.True(t, s.Items[3].Lines[0].Items[0].InlineStyle.SRTUnderline)
	assert.Nil(t, s.Items[4].Lines[0].Items[0].InlineStyle)
	assert.Nil(t, s.Items[5].Lines[0].Items[0].InlineStyle)
	assert.Nil(t, s.Items[5].Lines[1].Items[0].InlineStyle.SRTColor)
	assert.False(t, s.Items[5].Lines[1].Items[0].InlineStyle.SRTBold)
	assert.True(t, s.Items[5].Lines[1].Items[0].InlineStyle.SRTItalics)
	assert.False(t, s.Items[5].Lines[1].Items[0].InlineStyle.SRTUnderline)
	assert.True(t, s.Items[6].Lines[0].Items[0].InlineStyle.SRTItalics)
	assert.False(t, s.Items[6].Lines[0].Items[0].InlineStyle.SRTUnderline)
	assert.False(t, s.Items[6].Lines[0].Items[0].InlineStyle.SRTBold)
	assert.Nil(t, s.Items[6].Lines[0].Items[0].InlineStyle.SRTColor)
	assert.True(t, s.Items[6].Lines[1].Items[0].InlineStyle.SRTItalics)
	assert.False(t, s.Items[6].Lines[1].Items[0].InlineStyle.SRTUnderline)
	assert.False(t, s.Items[6].Lines[1].Items[0].InlineStyle.SRTBold)
	assert.Nil(t, s.Items[6].Lines[1].Items[0].InlineStyle.SRTColor)
	assert.True(t, s.Items[7].Lines[0].Items[0].InlineStyle.SRTItalics)
	assert.False(t, s.Items[7].Lines[0].Items[0].InlineStyle.SRTUnderline)
	assert.False(t, s.Items[7].Lines[0].Items[0].InlineStyle.SRTBold)
	assert.Nil(t, s.Items[7].Lines[0].Items[0].InlineStyle.SRTColor)
	assert.Nil(t, s.Items[8].Lines[0].Items[0].InlineStyle)
	assert.True(t, s.Items[9].Lines[0].Items[0].InlineStyle.SRTItalics)
	assert.Nil(t, s.Items[9].Lines[0].Items[1].InlineStyle)
	assert.True(t, s.Items[9].Lines[0].Items[2].InlineStyle.SRTItalics)
	assert.Nil(t, s.Items[9].Lines[0].Items[3].InlineStyle)

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

func TestSRTOverrideTags(t *testing.T) {
	testData := `1
00:00:01,000 --> 00:00:02,000
{\an8}A sign at the top

2
00:00:03,000 --> 00:00:04,000
{\an1}A sign at the {\i1}bottom left

3
00:00:05,000 --> 00:00:06,000
{\pos(400,570)}Positioning is dropped

4
00:00:07,000 --> 00:00:08,000
Curly {braces} are not tags`

	s, err := astisub.ReadFromSRT(strings.NewReader(testData))
	require.NoError(t, err)
	require.Len(t, s.Items, 4)

	// Alignment applies to the whole item
	assert.Equal(t, "A sign at the top", s.Items[0].Lines[0].String())
	assert.Equal(t, byte(8), s.Items[0].InlineStyle.SRTPosition)
	assert.Equal(t, "10%", s.Items[0].InlineStyle.WebVTTLine)
	assert.Equal(t, "", s.Items[0].InlineStyle.WebVTTAlign)
	assert.Nil(t, s.Items[0].Lines[0].Items[0].InlineStyle)

	// Text level tags only style what follows them
	assert.Equal(t, byte(1), s.Items[1].InlineStyle.SRTPosition)
	assert.Equal(t, "90%", s.Items[1].InlineStyle.WebVTTLine)
	assert.Equal(t, "left", s.Items[1].InlineStyle.WebVTTAlign)
	require.Len(t, s.Items[1].Lines[0].Items, 2)
	assert.Equal(t, "A sign at the ", s.Items[1].Lines[0].Items[0].Text)
	assert.Nil(t, s.Items[1].Lines[0].Items[0].InlineStyle)
	assert.Equal(t, "bottom left", s.Items[1].Lines[0].Items[1].Text)
	assert.True(t, s.Items[1].Lines[0].Items[1].InlineStyle.SRTItalics)

	// Tags that have no equivalent are dropped instead of being displayed
	assert.Equal(t, "Positioning is dropped", s.Items[2].Lines[0].String())
	assert.Nil(t, s.Items[2].InlineStyle)

	// Braces are only special when they hold tags
	assert.Equal(t, "Curly {braces} are not tags", s.Items[3].Lines[0].String())

	// Write to srt
	w := &bytes.Buffer{}
	err = s.WriteToSRT(w)
	assert.NoError(t, err)
	assert.Equal(t, `1
00:00:01,000 --> 00:00:02,000
{\an8}A sign at the top

2
00:00:03,000 --> 00:00:04,000
{\an1}A sign at the <i>bottom left</i>

3
00:00:05,000 --> 00:00:06,000
Positioning is dropped

4
00:00:07,000 --> 00:00:08,000
Curly {braces} are not tags
`, strings.TrimPrefix(w.String(), string(astisub.BytesBOM)))

	// Write to WebVTT
	w = &bytes.Buffer{}
	err = s.WriteToWebVTT(w)
	assert.NoError(t, err)
	assert.Equal(t, `WEBVTT

1
00:00:01.000 --> 00:00:02.000 line:10%
A sign at the top

2
00:00:03.000 --> 00:00:04.000 align:left line:90%
A sign at the <i>bottom left</i>

3
00:00:05.000 --> 00:00:06.000
Positioning is dropped

4
00:00:07.000 --> 00:00:08.000
Curly {braces} are not tags
`, w.String())
}

func TestSRTMissingEndTimeBoundary(t *testing.T) {
	testData := `1
00:48:52,500 --> 
Some subtitle text`

	_, err := astisub.ReadFromSRT(strings.NewReader(testData))
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing srt end time boundary")
}

func TestSRTParseDuration(t *testing.T) {
	testData := `
	1
	00:00:01.876-->00:0:03.390
	Duration without enclosing space
	
	2
	00:00:04:609-->00:0:05:985
	Duration without colon milliseconds`

	s, err := astisub.ReadFromSRT(strings.NewReader(testData))
	require.NoError(t, err)

	require.Len(t, s.Items, 2)
	assert.Equal(t, 1*time.Second+876*time.Millisecond, s.Items[0].StartAt)
	assert.Equal(t, 3*time.Second+390*time.Millisecond, s.Items[0].EndAt)
	assert.Equal(t, "Duration without enclosing space", s.Items[0].Lines[0].String())

	assert.Equal(t, 4*time.Second+609*time.Millisecond, s.Items[1].StartAt)
	assert.Equal(t, 5*time.Second+985*time.Millisecond, s.Items[1].EndAt)
	assert.Equal(t, "Duration without colon milliseconds", s.Items[1].Lines[0].String())
}
