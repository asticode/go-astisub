package astisub_test

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/asticode/go-astisub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebVTT(t *testing.T) {
	// Open
	s, err := astisub.OpenFile("./testdata/example-in.vtt")
	assert.NoError(t, err)
	assertSubtitleItems(t, s)
	// Comments
	assert.Equal(t, []string{"this a nice example", "of a VTT"}, s.Items[0].Comments)
	assert.Equal(t, []string{"This a comment inside the VTT", "and this is the second line"}, s.Items[1].Comments)
	// Regions
	assert.Equal(t, 2, len(s.Regions))
	assert.Equal(t, astisub.Region{ID: "fred", InlineStyle: &astisub.StyleAttributes{WebVTTLines: 3, WebVTTRegionAnchor: "0%,100%", WebVTTScroll: "up", WebVTTViewportAnchor: "10%,90%", WebVTTWidth: "40%"}}, *s.Regions["fred"])
	assert.Equal(t, astisub.Region{ID: "bill", InlineStyle: &astisub.StyleAttributes{WebVTTLines: 3, WebVTTRegionAnchor: "100%,100%", WebVTTScroll: "up", WebVTTViewportAnchor: "90%,90%", WebVTTWidth: "40%"}}, *s.Regions["bill"])
	assert.Equal(t, s.Regions["bill"], s.Items[0].Region)
	assert.Equal(t, s.Regions["fred"], s.Items[1].Region)
	// Styles
	assert.Equal(t, astisub.StyleAttributes{WebVTTAlign: "left", WebVTTPosition: "10%,start", WebVTTSize: "35%"}, *s.Items[1].InlineStyle)

	// No subtitles to write
	w := &bytes.Buffer{}
	err = astisub.Subtitles{}.WriteToWebVTT(w)
	assert.EqualError(t, err, astisub.ErrNoSubtitlesToWrite.Error())

	// Write
	c, err := ioutil.ReadFile("./testdata/example-out.vtt")
	assert.NoError(t, err)
	err = s.WriteToWebVTT(w)
	assert.NoError(t, err)
	assert.Equal(t, string(c), w.String())
}

func TestBroken1WebVTT(t *testing.T) {
	// Open bad, broken WebVTT file
	_, err := astisub.OpenFile("./testdata/broken-1-in.vtt")
	assert.Nil(t, err)
}

func TestNonUTF8WebVTT(t *testing.T) {
	_, err := astisub.OpenFile("./testdata/example-in-non-utf8.vtt")
	assert.Error(t, err)
}

func TestWebVTTWithVoiceName(t *testing.T) {
	testData := `WEBVTT

	NOTE this a example with voicename

	1
	00:02:34.000 --> 00:02:35.000
	<v.first.local Roger Bingham>I'm the fist speaker

	2
	00:02:34.000 --> 00:02:35.000
	<v Bingham>I'm the second speaker

	3
	00:00:04.000 --> 00:00:08.000
	<v Lee>What are you doing here?</v>

	4
	00:00:04.000 --> 00:00:08.000
	<v Bob>Incorrect tag?</vi>`

	s, err := astisub.ReadFromWebVTT(strings.NewReader(testData))
	assert.NoError(t, err)

	assert.Len(t, s.Items, 4)
	assert.Equal(t, "Roger Bingham", s.Items[0].Lines[0].VoiceName)
	assert.Equal(t, "Bingham", s.Items[1].Lines[0].VoiceName)
	assert.Equal(t, "Lee", s.Items[2].Lines[0].VoiceName)
	assert.Equal(t, "Bob", s.Items[3].Lines[0].VoiceName)

	b := &bytes.Buffer{}
	err = s.WriteToWebVTT(b)
	assert.NoError(t, err)
	assert.Equal(t, `WEBVTT

NOTE this a example with voicename

1
00:02:34.000 --> 00:02:35.000
<v Roger Bingham>I'm the fist speaker

2
00:02:34.000 --> 00:02:35.000
<v Bingham>I'm the second speaker

3
00:00:04.000 --> 00:00:08.000
<v Lee>What are you doing here?

4
00:00:04.000 --> 00:00:08.000
<v Bob>Incorrect tag?
`, b.String())
}

func TestWebVTTWithTimestampMap(t *testing.T) {
	testData := `WEBVTT
	X-TIMESTAMP-MAP=MPEGTS:180000, LOCAL:00:00:00.000

	00:00.933 --> 00:02.366
	♪ ♪

	00:02.400 --> 00:03.633
	Evening.`

	s, err := astisub.ReadFromWebVTT(strings.NewReader(testData))
	assert.NoError(t, err)

	assert.Len(t, s.Items, 2)

	assert.Equal(t, s.Items[0].StartAt.Milliseconds(), int64(933))
	assert.Equal(t, s.Items[0].EndAt.Milliseconds(), int64(2366))
	assert.Equal(t, s.Items[1].StartAt.Milliseconds(), int64(2400))
	assert.Equal(t, s.Items[1].EndAt.Milliseconds(), int64(3633))
	assert.Equal(t, s.Metadata.WebVTTTimestampMap.Offset(), time.Duration(time.Second*2))

	b := &bytes.Buffer{}
	err = s.WriteToWebVTT(b)
	assert.NoError(t, err)
	assert.Equal(t, `WEBVTT
X-TIMESTAMP-MAP=LOCAL:00:00:00.000,MPEGTS:180000

1
00:00:00.933 --> 00:00:02.366
♪ ♪

2
00:00:02.400 --> 00:00:03.633
Evening.
`, b.String())
}

func TestWebVTTTags(t *testing.T) {
	testData := `WEBVTT

	00:01:00.000 --> 00:02:00.000
	<u><i>Italic with underline text</i></u> some extra

	00:02:00.000 --> 00:03:00.000
	<lang en>English here</lang> <c.yellow.bg_blue>Yellow text on blue background</c>

	00:03:00.000 --> 00:04:00.000
	<v Joe><c.red><i>Joe's words are red in italic</i></c>

	00:04:00.000 --> 00:05:00.000
	<customed_tag.class1.class2>Text here</customed_tag>

	00:05:00.000 --> 00:06:00.000
	<v Joe>Joe says something</v> <v Bob>Bob says something</v>

	00:06:00.000 --> 00:07:00.000
	Text with a <00:06:30.000>timestamp in the middle

	00:08:00.000 --> 00:09:00.000
	<i>Test with multi line italics
	Terminated on the next line</i>
	
	00:09:00.000 --> 00:10:00.000
	<i>Unterminated styles
	
	00:10:00.000 --> 00:11:00.000
	Do no fall to the next item
	
	00:12:00.000 --> 00:13:00.000
	<i>x</i>^3 * <i>x</i> = 100`

	s, err := astisub.ReadFromWebVTT(strings.NewReader(testData))
	require.NoError(t, err)

	require.Len(t, s.Items, 10)

	b := &bytes.Buffer{}
	err = s.WriteToWebVTT(b)
	require.NoError(t, err)
	require.Equal(t, `WEBVTT

1
00:01:00.000 --> 00:02:00.000
<u><i>Italic with underline text</i></u> some extra

2
00:02:00.000 --> 00:03:00.000
<lang en>English here</lang> <c.yellow.bg_blue>Yellow text on blue background</c>

3
00:03:00.000 --> 00:04:00.000
<v Joe><c.red><i>Joe's words are red in italic</i></c>

4
00:04:00.000 --> 00:05:00.000
<customed_tag.class1.class2>Text here</customed_tag>

5
00:05:00.000 --> 00:06:00.000
<v Joe>Joe says something Bob says something

6
00:06:00.000 --> 00:07:00.000
Text with a <00:06:30.000>timestamp in the middle

7
00:08:00.000 --> 00:09:00.000
<i>Test with multi line italics</i>
<i>Terminated on the next line</i>

8
00:09:00.000 --> 00:10:00.000
<i>Unterminated styles</i>

9
00:10:00.000 --> 00:11:00.000
Do no fall to the next item

10
00:12:00.000 --> 00:13:00.000
<i>x</i>^3 * <i>x</i> = 100
`, b.String())
}

func TestWebVTTParseDuration(t *testing.T) {
	testData := `WEBVTT
	1
	00:00:01.876-->00:0:03.390
	Duration without enclosing space

	2
	00:00:03.391-->00:00:06.567	align:middle
	Duration with tab spaced styles`

	s, err := astisub.ReadFromWebVTT(strings.NewReader(testData))
	require.NoError(t, err)

	require.Len(t, s.Items, 2)
	assert.Equal(t, 1*time.Second+876*time.Millisecond, s.Items[0].StartAt)
	assert.Equal(t, 3*time.Second+390*time.Millisecond, s.Items[0].EndAt)
	assert.Equal(t, "Duration without enclosing space", s.Items[0].Lines[0].String())
	assert.Equal(t, 3*time.Second+391*time.Millisecond, s.Items[1].StartAt)
	assert.Equal(t, 6*time.Second+567*time.Millisecond, s.Items[1].EndAt)
	assert.Equal(t, "Duration with tab spaced styles", s.Items[1].Lines[0].String())
	assert.NotNil(t, s.Items[1].InlineStyle)
	assert.Equal(t, s.Items[1].InlineStyle.WebVTTAlign, "middle")
}
