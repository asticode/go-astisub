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
	s, err := astisub.OpenFile("./testdata/example-in.ttml")
	assert.NoError(t, err)
	assertSubtitleItems(t, s)
	// Metadata
	assert.Equal(t, &astisub.Metadata{Framerate: 25, Language: astisub.LanguageFrench, Title: "Title test", TTMLCopyright: "Copyright test"}, s.Metadata)
	// Styles
	assert.Equal(t, 3, len(s.Styles))
	assert.Equal(t, astisub.Style{ID: "style_0", InlineStyle: &astisub.StyleAttributes{TTMLColor: "white", TTMLExtent: "100% 10%", TTMLFontFamily: "sansSerif", TTMLFontStyle: "normal", TTMLOrigin: "0% 90%", TTMLTextAlign: "center"}, Style: s.Styles["style_2"]}, *s.Styles["style_0"])
	assert.Equal(t, astisub.Style{ID: "style_1", InlineStyle: &astisub.StyleAttributes{TTMLColor: "white", TTMLExtent: "100% 13%", TTMLFontFamily: "sansSerif", TTMLFontStyle: "normal", TTMLOrigin: "0% 87%", TTMLTextAlign: "center"}}, *s.Styles["style_1"])
	assert.Equal(t, astisub.Style{ID: "style_2", InlineStyle: &astisub.StyleAttributes{TTMLColor: "white", TTMLExtent: "100% 20%", TTMLFontFamily: "sansSerif", TTMLFontStyle: "normal", TTMLOrigin: "0% 80%", TTMLTextAlign: "center"}}, *s.Styles["style_2"])
	// Regions
	assert.Equal(t, 3, len(s.Regions))
	assert.Equal(t, astisub.Region{ID: "region_0", Style: s.Styles["style_0"], InlineStyle: &astisub.StyleAttributes{TTMLColor: "blue"}}, *s.Regions["region_0"])
	assert.Equal(t, astisub.Region{ID: "region_1", Style: s.Styles["style_1"], InlineStyle: &astisub.StyleAttributes{}}, *s.Regions["region_1"])
	assert.Equal(t, astisub.Region{ID: "region_2", Style: s.Styles["style_2"], InlineStyle: &astisub.StyleAttributes{}}, *s.Regions["region_2"])
	// Items
	assert.Equal(t, s.Regions["region_1"], s.Items[0].Region)
	assert.Equal(t, s.Styles["style_1"], s.Items[0].Style)
	assert.Equal(t, &astisub.StyleAttributes{TTMLColor: "red"}, s.Items[0].InlineStyle)
	assert.Equal(t, []astisub.Line{{Items: []astisub.LineItem{{Style: s.Styles["style_1"], InlineStyle: &astisub.StyleAttributes{TTMLColor: "black"}, Text: "(deep rumbling)"}}}}, s.Items[0].Lines)
	assert.Equal(t, []astisub.Line{{Items: []astisub.LineItem{{InlineStyle: &astisub.StyleAttributes{}, Text: "MAN:"}}}, {Items: []astisub.LineItem{{InlineStyle: &astisub.StyleAttributes{}, Text: "How did we"}, {InlineStyle: &astisub.StyleAttributes{TTMLColor: "green"}, Style: s.Styles["style_1"], Text: "end up"}, {InlineStyle: &astisub.StyleAttributes{}, Text: "here?"}}}}, s.Items[1].Lines)
	assert.Equal(t, []astisub.Line{{Items: []astisub.LineItem{{InlineStyle: &astisub.StyleAttributes{}, Style: s.Styles["style_1"], Text: "This place is horrible."}}}}, s.Items[2].Lines)
	assert.Equal(t, []astisub.Line{{Items: []astisub.LineItem{{InlineStyle: &astisub.StyleAttributes{}, Style: s.Styles["style_1"], Text: "Smells like balls."}}}}, s.Items[3].Lines)
	assert.Equal(t, []astisub.Line{{Items: []astisub.LineItem{{InlineStyle: &astisub.StyleAttributes{}, Style: s.Styles["style_2"], Text: "We don't belong"}}}, {Items: []astisub.LineItem{{InlineStyle: &astisub.StyleAttributes{}, Style: s.Styles["style_1"], Text: "in this shithole."}}}}, s.Items[4].Lines)
	assert.Equal(t, []astisub.Line{{Items: []astisub.LineItem{{InlineStyle: &astisub.StyleAttributes{}, Style: s.Styles["style_2"], Text: "(computer playing"}}}, {Items: []astisub.LineItem{{InlineStyle: &astisub.StyleAttributes{}, Style: s.Styles["style_1"], Text: "electronic melody)"}}}}, s.Items[5].Lines)

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
