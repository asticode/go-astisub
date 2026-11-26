package astisub_test

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/asticode/go-astikit"

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
	assert.Equal(t, astisub.Style{ID: "style_0", InlineStyle: &astisub.StyleAttributes{STLColor: astisub.ColorWhite, TTMLColor: astisub.ColorWhite, TTMLExtent: astikit.StrPtr("100% 10%"), TTMLFontFamily: astikit.StrPtr("sansSerif"), TTMLFontStyle: astikit.StrPtr("normal"), TTMLOrigin: astikit.StrPtr("0% 90%"), TTMLTextAlign: astikit.StrPtr("center"), WebVTTAlign: "center", WebVTTLine: "0%", WebVTTLines: 2, WebVTTPosition: &astisub.WebVTTPosition{XPosition: "90%"}, WebVTTRegionAnchor: "0%,100%", WebVTTScroll: "", WebVTTSize: "100%", WebVTTViewportAnchor: "0%,100%", WebVTTWidth: "100%"}, Style: s.Styles["style_2"]}, *s.Styles["style_0"])
	assert.Equal(t, astisub.Style{ID: "style_1", InlineStyle: &astisub.StyleAttributes{STLColor: astisub.ColorWhite, TTMLColor: astisub.ColorWhite, TTMLExtent: astikit.StrPtr("100% 13%"), TTMLFontFamily: astikit.StrPtr("sansSerif"), TTMLFontStyle: astikit.StrPtr("normal"), TTMLOrigin: astikit.StrPtr("0% 87%"), TTMLTextAlign: astikit.StrPtr("center"), WebVTTAlign: "center", WebVTTLine: "0%", WebVTTLines: 2, WebVTTPosition: &astisub.WebVTTPosition{XPosition: "87%"}, WebVTTRegionAnchor: "0%,100%", WebVTTScroll: "", WebVTTSize: "100%", WebVTTViewportAnchor: "0%,100%", WebVTTWidth: "100%"}}, *s.Styles["style_1"])
	assert.Equal(t, astisub.Style{ID: "style_2", InlineStyle: &astisub.StyleAttributes{STLColor: astisub.ColorWhite, TTMLColor: astisub.ColorWhite, TTMLExtent: astikit.StrPtr("100% 20%"), TTMLFontFamily: astikit.StrPtr("sansSerif"), TTMLFontStyle: astikit.StrPtr("normal"), TTMLOrigin: astikit.StrPtr("0% 80%"), TTMLTextAlign: astikit.StrPtr("center"), WebVTTAlign: "center", WebVTTLine: "0%", WebVTTLines: 4, WebVTTPosition: &astisub.WebVTTPosition{XPosition: "80%"}, WebVTTRegionAnchor: "0%,100%", WebVTTScroll: "", WebVTTSize: "100%", WebVTTViewportAnchor: "0%,100%", WebVTTWidth: "100%"}}, *s.Styles["style_2"])
	// Regions
	assert.Equal(t, 3, len(s.Regions))
	assert.Equal(t, astisub.Region{ID: "region_0", Style: s.Styles["style_0"], InlineStyle: &astisub.StyleAttributes{STLColor: astisub.ColorBlue, TTMLColor: astisub.ColorBlue}}, *s.Regions["region_0"])
	assert.Equal(t, astisub.Region{ID: "region_1", Style: s.Styles["style_1"], InlineStyle: &astisub.StyleAttributes{}}, *s.Regions["region_1"])
	assert.Equal(t, astisub.Region{ID: "region_2", Style: s.Styles["style_2"], InlineStyle: &astisub.StyleAttributes{}}, *s.Regions["region_2"])
	// Items
	assert.Equal(t, s.Regions["region_1"], s.Items[0].Region)
	assert.Equal(t, s.Styles["style_1"], s.Items[0].Style)
	assert.Equal(t, &astisub.StyleAttributes{STLColor: astisub.ColorRed, TTMLColor: astisub.ColorRed}, s.Items[0].InlineStyle)
	assert.Equal(t, []astisub.Line{{Items: []astisub.LineItem{{Style: s.Styles["style_1"], InlineStyle: &astisub.StyleAttributes{STLColor: astisub.ColorBlack, TTMLColor: astisub.ColorBlack}, Text: "(deep rumbling)"}}}}, s.Items[0].Lines)
	assert.Equal(t, []astisub.Line{{Items: []astisub.LineItem{{InlineStyle: &astisub.StyleAttributes{}, Text: "MAN:"}}}, {Items: []astisub.LineItem{{InlineStyle: &astisub.StyleAttributes{}, Text: "How did we "}, {InlineStyle: &astisub.StyleAttributes{STLColor: astisub.ColorGreen, TTMLColor: astisub.ColorGreen}, Style: s.Styles["style_1"], Text: "end up"}, {InlineStyle: &astisub.StyleAttributes{}, Text: " here?"}}}}, s.Items[1].Lines)
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

func TestTTMLBreakLines(t *testing.T) {
	// Open
	s, err := astisub.OpenFile("./testdata/example-in-breaklines.ttml")
	assert.NoError(t, err)

	// Write
	w := &bytes.Buffer{}
	err = s.WriteToTTML(w)
	assert.NoError(t, err)

	c, err := ioutil.ReadFile("./testdata/example-out-breaklines.ttml")
	assert.NoError(t, err)

	assert.Equal(t, strings.TrimSpace(string(c)), strings.TrimSpace(w.String()))
}

func TestWriteToTTMLWithIndentOption(t *testing.T) {
	// Open
	s, err := astisub.OpenFile("./testdata/example-in.ttml")
	assert.NoError(t, err)

	// Write
	w := &bytes.Buffer{}

	err = s.WriteToTTML(w, astisub.WriteToTTMLWithIndentOption(""))
	assert.NoError(t, err)

	c, err := ioutil.ReadFile("./testdata/example-out-no-indent.ttml")
	assert.NoError(t, err)

	assert.Equal(t, strings.TrimSpace(string(c)), strings.TrimSpace(w.String()))
}

func TestTTMLMergeStyleAttributes(t *testing.T) {
	// Open
	s, err := astisub.OpenFile("./testdata/example-in-merging-style.ttml")
	assert.NoError(t, err)
	// Styles
	assert.Equal(t, 4, len(s.Items))
	assert.Equal(t, s.Items[0].Region.ID, "region_0")
	assert.Equal(t, s.Items[0].Style.ID, "style_1")
	assert.Equal(t, s.Items[0].InlineStyle.TTMLColor, astisub.ColorRed)
	assert.Equal(t, s.Items[1].Region.ID, "region_1")
	assert.Equal(t, s.Items[1].Style.ID, "style_0")
	assert.Equal(t, s.Items[1].Lines[0].Items[0].Style.ID, "style_1")
	assert.Equal(t, s.Items[2].Region.ID, "region_2")
	assert.Equal(t, s.Items[2].Style.ID, "style_0")
	assert.Equal(t, s.Items[3].InlineStyle.TTMLColor, astisub.ColorBlue)
}

func TestTTMLTiming(t *testing.T) {
	s, err := astisub.ReadFromTTML(strings.NewReader(`<tt xmlns="http://www.w3.org/ns/ttml"><body>
<div><p begin="00:00:01.000" dur="2s">Dur</p></div>
<div begin="00:00:05.000" end="00:00:07.000"><p>From the div</p></div>
</body></tt>`))
	assert.NoError(t, err)
	assert.Equal(t, 2, len(s.Items))
	assert.Equal(t, []time.Duration{time.Second, 3 * time.Second}, []time.Duration{s.Items[0].StartAt, s.Items[0].EndAt})
	assert.Equal(t, []time.Duration{5 * time.Second, 7 * time.Second}, []time.Duration{s.Items[1].StartAt, s.Items[1].EndAt})

	_, err = astisub.ReadFromTTML(strings.NewReader(`<tt xmlns="http://www.w3.org/ns/ttml"><body><div><p begin="1s">No end</p></div></body></tt>`))
	assert.EqualError(t, err, `astisub: subtitle "" has no begin or end time`)
}

func TestTTMLStylesSharingAParent(t *testing.T) {
	s, err := astisub.ReadFromTTML(strings.NewReader(`<tt xmlns="http://www.w3.org/ns/ttml" xmlns:tts="http://www.w3.org/ns/ttml#styling"><head><styling>
<style xml:id="base" tts:color="white"/>
<style xml:id="italic" style="base" tts:fontStyle="italic"/>
<style xml:id="bold" style="base" tts:fontWeight="bold"/>
<style xml:id="plain"/>
</styling></head><body><div><p begin="00:00:01.000" end="00:00:02.000" style="italic">a</p></div></body></tt>`))
	assert.NoError(t, err)
	for _, id := range []string{"italic", "bold"} {
		if assert.NotNil(t, s.Styles[id].Style, id) {
			assert.Equal(t, "base", s.Styles[id].Style.ID)
		}
	}
	assert.Nil(t, s.Styles["base"].Style)
	assert.Nil(t, s.Styles["plain"].Style)

	_, err = astisub.ReadFromTTML(strings.NewReader(`<tt xmlns="http://www.w3.org/ns/ttml"><head><styling>
<style xml:id="a" style="missing"/>
</styling></head><body/></tt>`))
	assert.EqualError(t, err, "astisub: Style missing requested by style a doesn't exist")
}
