package astisub_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/asticode/go-astisub"
	"github.com/stretchr/testify/assert"
)

func TestTTML(t *testing.T) {
	// Open
	s, err := astisub.Open("./testdata/example-in.ttml")
	assert.NoError(t, err)
	assertSubtitleItems(t, s)
	// Metadata
	assert.Equal(t, &astisub.Metadata{Copyright: "Copyright test", Framerate: 25, Language: astisub.LanguageFrench, Title: "Title test"}, s.Metadata)
	// Styles
	assert.Equal(t, 3, len(s.Styles))
	assert.Equal(t, astisub.Style{ID: "style_0", InlineStyle: &astisub.StyleAttributes{Color: "white", Extent: "100% 10%", FontFamily: "sansSerif", FontStyle: "normal", Origin: "0% 90%", TextAlign: "center"}, Style: s.Styles[2]}, *s.Styles[0])
	assert.Equal(t, astisub.Style{ID: "style_1", InlineStyle: &astisub.StyleAttributes{Color: "white", Extent: "100% 13%", FontFamily: "sansSerif", FontStyle: "normal", Origin: "0% 87%", TextAlign: "center"}}, *s.Styles[1])
	assert.Equal(t, astisub.Style{ID: "style_2", InlineStyle: &astisub.StyleAttributes{Color: "white", Extent: "100% 20%", FontFamily: "sansSerif", FontStyle: "normal", Origin: "0% 80%", TextAlign: "center"}}, *s.Styles[2])
	// Regions
	assert.Equal(t, 3, len(s.Regions))
	assert.Equal(t, astisub.Region{ID: "region_0", Style: s.Styles[0], InlineStyle: &astisub.StyleAttributes{Color: "blue"}}, *s.Regions[0])
	assert.Equal(t, astisub.Region{ID: "region_1", Style: s.Styles[1], InlineStyle: &astisub.StyleAttributes{}}, *s.Regions[1])
	assert.Equal(t, astisub.Region{ID: "region_2", Style: s.Styles[2], InlineStyle: &astisub.StyleAttributes{}}, *s.Regions[2])
	// Items
	assert.Equal(t, s.Regions[1], s.Items[0].Region)
	assert.Equal(t, s.Styles[1], s.Items[0].Style)
	assert.Equal(t, &astisub.StyleAttributes{Color: "red"}, s.Items[0].InlineStyle)
	assert.Equal(t, []astisub.Line{{{Style: s.Styles[1], InlineStyle: &astisub.StyleAttributes{Color: "black"}, Text: "(deep rumbling)"}}}, s.Items[0].Lines)
	assert.Equal(t, []astisub.Line{{{InlineStyle: &astisub.StyleAttributes{}, Text: "MAN:"}}, {{InlineStyle: &astisub.StyleAttributes{}, Text: "How did we"}, {InlineStyle: &astisub.StyleAttributes{Color: "green"}, Style: s.Styles[1], Text: "end up"}, {InlineStyle: &astisub.StyleAttributes{}, Text: "here?"}}}, s.Items[1].Lines)
	assert.Equal(t, []astisub.Line{{{InlineStyle: &astisub.StyleAttributes{}, Style: s.Styles[1], Text: "This place is horrible."}}}, s.Items[2].Lines)
	assert.Equal(t, []astisub.Line{{{InlineStyle: &astisub.StyleAttributes{}, Style: s.Styles[1], Text: "Smells like balls."}}}, s.Items[3].Lines)
	assert.Equal(t, []astisub.Line{{{InlineStyle: &astisub.StyleAttributes{}, Style: s.Styles[2], Text: "We don't belong"}}, {{InlineStyle: &astisub.StyleAttributes{}, Style: s.Styles[1], Text: "in this shithole."}}}, s.Items[4].Lines)
	assert.Equal(t, []astisub.Line{{{InlineStyle: &astisub.StyleAttributes{}, Style: s.Styles[2], Text: "(computer playing"}}, {{InlineStyle: &astisub.StyleAttributes{}, Style: s.Styles[1], Text: "electronic melody)"}}}, s.Items[5].Lines)

	// No subtitles to write
	w := &bytes.Buffer{}
	err = astisub.Subtitles{}.WriteToTTML(w)
	assert.EqualError(t, err, astisub.ErrNoSubtitlesToWrite.Error())

	// Write
	c, err := ioutil.ReadFile("./testdata/example-out.ttml")
	assert.NoError(t, err)
	err = s.WriteToTTML(w)
	assert.NoError(t, err)
	assert.Equal(t, string(c), w.String())
}
