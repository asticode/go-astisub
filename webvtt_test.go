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
