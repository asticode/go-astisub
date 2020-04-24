package astisub_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/asticode/go-astisub"
	"github.com/stretchr/testify/assert"
)

func TestSRT_Open(t *testing.T) {
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

func TestSRT_FromLines(t *testing.T) {
	f, err := os.Open("./testdata/example-in.srt")
	assert.NoError(t, err)
	defer f.Close()

	lines, err := astisub.ScanLines(f)
	assert.NoError(t, err)

	subs := astisub.NewSubtitles()
	err = astisub.ParseFromSRTLines(subs, lines)
	assert.NoError(t, err)
	assertSubtitleItems(t, subs)
}
