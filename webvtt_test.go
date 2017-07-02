package astisub_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/asticode/go-astisub"
	"github.com/stretchr/testify/assert"
)

func TestWebVTT(t *testing.T) {
	// Open
	s, err := astisub.Open("./testdata/example-in.vtt")
	assert.NoError(t, err)
	assertSubtitleItems(t, s)
	// Comments
	assert.Equal(t, []string{"this a nice example", "of a VTT"}, s.Items[0].Comments)
	assert.Equal(t, []string{"This a comment inside the VTT", "and this is the second line"}, s.Items[1].Comments)
	// Regions
	assert.Equal(t, 2, len(s.Regions))
	assert.Equal(t, astisub.Region{ID: "fred", InlineStyle: &astisub.StyleAttributes{Lines: 3, RegionAnchor: "0%,100%", Scroll: "up", ViewportAnchor: "10%,90%", Width: "40%"}}, *s.Regions[0])
	assert.Equal(t, astisub.Region{ID: "bill", InlineStyle: &astisub.StyleAttributes{Lines: 3, RegionAnchor: "100%,100%", Scroll: "up", ViewportAnchor: "90%,90%", Width: "40%"}}, *s.Regions[1])
	assert.Equal(t, s.Regions[1], s.Items[0].Region)
	assert.Equal(t, s.Regions[0], s.Items[1].Region)
	// Styles
	assert.Equal(t, astisub.StyleAttributes{Align: "left", Position: "10%,start", Size: "35%"}, *s.Items[1].InlineStyle)

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
