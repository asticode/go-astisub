package astisub_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

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
	s, err := astisub.OpenFile("./testdata/example-styled-in.srt")
	assert.NoError(t, err)
	assertStyledSubtitleItems(t, s)
	assertSRTSubtitleStyles(t, s)

	// No subtitles to write
	w := &bytes.Buffer{}
	err = astisub.Subtitles{}.WriteToSRT(w)
	assert.EqualError(t, err, astisub.ErrNoSubtitlesToWrite.Error())

	// Write
	c, err := os.ReadFile("./testdata/example-styled-out.srt")
	assert.NoError(t, err)
	err = s.WriteToSRT(w)
	assert.NoError(t, err)
	assert.Equal(t, string(c), w.String())
}

func TestReadSRTWriteWebVTTStyled(t *testing.T) {
	// Open
	s, err := astisub.OpenFile("./testdata/example-styled-in.srt")
	assert.NoError(t, err)
	assertStyledSubtitleItems(t, s)
	assertSRTSubtitleStyles(t, s)

	w := &bytes.Buffer{}

	// Write
	c, err := os.ReadFile("./testdata/example-styled-out.vtt")
	assert.NoError(t, err)
	err = s.WriteToWebVTT(w)
	assert.NoError(t, err)
	assert.Equal(t, string(c), w.String())
}
