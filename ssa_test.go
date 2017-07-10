package astisub_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/asticode/go-astisub"
	"github.com/asticode/go-astitools/ptr"
	"github.com/stretchr/testify/assert"
)

func assertSSAStyle(t *testing.T, e, a astisub.Style) {
	assert.Equal(t, e.ID, a.ID)
	assertSSAStyleAttributes(t, *e.InlineStyle, *a.InlineStyle)
}

func assertSSAStyleAttributes(t *testing.T, e, a astisub.StyleAttributes) {
	if e.SSAAlignment != nil {
		assert.Equal(t, *e.SSAAlignment, *a.SSAAlignment)
	}
	if e.SSAAlphaLevel != nil {
		assert.Equal(t, *e.SSAAlphaLevel, *a.SSAAlphaLevel)
	}
	if e.SSABackColour != nil {
		assert.Equal(t, *e.SSABackColour, *a.SSABackColour)
	}
	if e.SSABold != nil {
		assert.Equal(t, *e.SSABold, *a.SSABold)
	}
	if e.SSABorderStyle != nil {
		assert.Equal(t, *e.SSABorderStyle, *a.SSABorderStyle)
	}
	if e.SSAFontSize != nil {
		assert.Equal(t, e.SSAFontName, a.SSAFontName)
	}
	if e.SSAFontSize != nil {
		assert.Equal(t, *e.SSAFontSize, *a.SSAFontSize)
	}
	if e.SSALayer != nil {
		assert.Equal(t, *e.SSALayer, *a.SSALayer)
	}
	if e.SSAMarked != nil {
		assert.Equal(t, *e.SSAMarked, *a.SSAMarked)
	}
	if e.SSAMarginLeft != nil {
		assert.Equal(t, *e.SSAMarginLeft, *a.SSAMarginLeft)
	}
	if e.SSAMarginRight != nil {
		assert.Equal(t, *e.SSAMarginRight, *a.SSAMarginRight)
	}
	if e.SSAMarginVertical != nil {
		assert.Equal(t, *e.SSAMarginVertical, *a.SSAMarginVertical)
	}
	if e.SSAOutline != nil {
		assert.Equal(t, *e.SSAOutline, *a.SSAOutline)
	}
	if e.SSAOutlineColour != nil {
		assert.Equal(t, *e.SSAOutlineColour, *a.SSAOutlineColour)
	}
	if e.SSAPrimaryColour != nil {
		assert.Equal(t, *e.SSAPrimaryColour, *a.SSAPrimaryColour)
	}
	if e.SSASecondaryColour != nil {
		assert.Equal(t, *e.SSASecondaryColour, *a.SSASecondaryColour)
	}
	if e.SSAShadow != nil {
		assert.Equal(t, *e.SSAShadow, *a.SSAShadow)
	}
}

func TestSSA(t *testing.T) {
	// Open
	s, err := astisub.OpenFile("./testdata/example-in.ssa")
	assert.NoError(t, err)
	assertSubtitleItems(t, s)
	// Metadata
	assert.Equal(t, &astisub.Metadata{Comments: []string{"Comment 1", "Comment 2"}, Copyright: "Copyright test", Title: "SSA test"}, s.Metadata)
	// Styles
	assert.Equal(t, 3, len(s.Styles))
	assertSSAStyle(t, astisub.Style{ID: "1", InlineStyle: &astisub.StyleAttributes{SSAAlignment: astiptr.Int(7), SSAAlphaLevel: astiptr.Float(0.1), SSABackColour: &astisub.Color{Alpha: 128, Red: 8}, SSABold: astiptr.Bool(true), SSABorderStyle: astiptr.Int(7), SSAFontName: "f1", SSAFontSize: astiptr.Float(4), SSAOutline: astiptr.Int(1), SSAOutlineColour: &astisub.Color{Green: 255, Red: 255}, SSAMarginLeft: astiptr.Int(1), SSAMarginRight: astiptr.Int(4), SSAMarginVertical: astiptr.Int(7), SSAPrimaryColour: &astisub.Color{Green: 255, Red: 255}, SSASecondaryColour: &astisub.Color{Green: 255, Red: 255}, SSAShadow: astiptr.Int(4)}}, *s.Styles["1"])
	assertSSAStyle(t, astisub.Style{ID: "2", InlineStyle: &astisub.StyleAttributes{SSAAlignment: astiptr.Int(8), SSAAlphaLevel: astiptr.Float(0.2), SSABackColour: &astisub.Color{Blue: 15, Green: 15, Red: 15}, SSABold: astiptr.Bool(true), SSABorderStyle: astiptr.Int(8), SSAEncoding: astiptr.Int(1), SSAFontName: "f2", SSAFontSize: astiptr.Float(5), SSAOutline: astiptr.Int(2), SSAOutlineColour: &astisub.Color{Green: 255, Red: 255}, SSAMarginLeft: astiptr.Int(2), SSAMarginRight: astiptr.Int(5), SSAMarginVertical: astiptr.Int(8), SSAPrimaryColour: &astisub.Color{Blue: 239, Green: 239, Red: 239}, SSASecondaryColour: &astisub.Color{Green: 255, Red: 255}, SSAShadow: astiptr.Int(5)}}, *s.Styles["2"])
	assertSSAStyle(t, astisub.Style{ID: "3", InlineStyle: &astisub.StyleAttributes{SSAAlignment: astiptr.Int(9), SSAAlphaLevel: astiptr.Float(0.3), SSABackColour: &astisub.Color{Red: 8}, SSABorderStyle: astiptr.Int(9), SSAEncoding: astiptr.Int(2), SSAFontName: "f3", SSAFontSize: astiptr.Float(6), SSAOutline: astiptr.Int(3), SSAOutlineColour: &astisub.Color{Red: 8}, SSAMarginLeft: astiptr.Int(3), SSAMarginRight: astiptr.Int(6), SSAMarginVertical: astiptr.Int(9), SSAPrimaryColour: &astisub.Color{Blue: 180, Green: 252, Red: 252}, SSASecondaryColour: &astisub.Color{Blue: 180, Green: 252, Red: 252}, SSAShadow: astiptr.Int(6)}}, *s.Styles["3"])
	// Items
	assertSSAStyleAttributes(t, astisub.StyleAttributes{SSAEffect: "test", SSAMarked: astiptr.Bool(false), SSAMarginLeft: astiptr.Int(1234), SSAMarginRight: astiptr.Int(2345), SSAMarginVertical: astiptr.Int(3456)}, *s.Items[0].InlineStyle)
	assert.Equal(t, s.Styles["1"], s.Items[0].Style)
	assert.Equal(t, []astisub.Line{{Items: []astisub.LineItem{{InlineStyle: &astisub.StyleAttributes{SSAEffect: "{\\pos(400,570)}"}, Text: "(deep rumbling)"}}, VoiceName: "Cher"}}, s.Items[0].Lines)
	assert.Equal(t, s.Styles["2"], s.Items[1].Style)
	assert.Equal(t, s.Styles["3"], s.Items[2].Style)
	assert.Equal(t, s.Styles["1"], s.Items[3].Style)
	assert.Equal(t, s.Styles["2"], s.Items[4].Style)
	assert.Equal(t, s.Styles["3"], s.Items[5].Style)

	// No subtitles to write
	w := &bytes.Buffer{}
	err = astisub.Subtitles{}.WriteToSSA(w)
	assert.EqualError(t, err, astisub.ErrNoSubtitlesToWrite.Error())

	// Write
	c, err := ioutil.ReadFile("./testdata/example-out.ssa")
	assert.NoError(t, err)
	err = s.WriteToSSA(w)
	assert.NoError(t, err)
	assert.Equal(t, string(c), w.String())
}
