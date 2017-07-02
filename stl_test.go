package astisub_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/asticode/go-astisub"
	"github.com/stretchr/testify/assert"
)

func TestSTL(t *testing.T) {
	// Open
	s, err := astisub.Open("./testdata/example-in.stl")
	assert.NoError(t, err)
	assertSubtitleItems(t, s)
	// Metadata
	assert.Equal(t, &astisub.Metadata{Copyright: "Copyright test", Framerate: 25, Language: astisub.LanguageFrench, Title: "Title test"}, s.Metadata)

	// No subtitles to write
	w := &bytes.Buffer{}
	err = astisub.Subtitles{}.WriteToSTL(w)
	assert.EqualError(t, err, astisub.ErrNoSubtitlesToWrite.Error())

	// Write
	c, err := ioutil.ReadFile("./testdata/example-out.stl")
	assert.NoError(t, err)
	err = s.WriteToSTL(w)
	assert.NoError(t, err)
	assert.Equal(t, string(c), w.String())
}
