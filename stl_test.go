package astisub_test

import (
	"bytes"
	"io/ioutil"
	"testing"
	"time"

	"github.com/asticode/go-astikit"
	"github.com/asticode/go-astisub"
	"github.com/stretchr/testify/assert"
)

func TestSTL(t *testing.T) {
	// Init
	creationDate, _ := time.Parse("060102", "170702")
	revisionDate, _ := time.Parse("060102", "010101")

	// Open
	s, err := astisub.OpenFile("./testdata/example-in.stl")
	assert.NoError(t, err)
	assertSubtitleItems(t, s)
	// Metadata
	assert.Equal(t, &astisub.Metadata{
		Framerate: 25,
		Language:  astisub.LanguageFrench,
		STLMaximumNumberOfDisplayableCharactersInAnyTextRow: astikit.IntPtr(40),
		STLMaximumNumberOfDisplayableRows:                   astikit.IntPtr(23),
		STLPublisher:                                        "Copyright test",
		STLDisplayStandardCode:                              "1",
		STLSubtitleListReferenceCode:                        "12345678",
		STLCountryOfOrigin:                                  "FRA",
		Title:                                               "Title test",
		CreationDate:                                        &creationDate,
		RevisionDate:                                        &revisionDate},
		s.Metadata)

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

func TestOPNSTL(t *testing.T) {
	// Init
	creationDate, _ := time.Parse("060102", "200110")
	revisionDate, _ := time.Parse("060102", "200110")

	// Open
	s, err := astisub.OpenFile("./testdata/example-opn-in.stl")
	assert.NoError(t, err)
	//assertSubtitleItems(t, s)
	// Metadata
	assert.Equal(t, &astisub.Metadata{
		Framerate: 25,
		Language:  astisub.LanguageEnglish,
		STLMaximumNumberOfDisplayableCharactersInAnyTextRow: astikit.IntPtr(38),
		STLMaximumNumberOfDisplayableRows:                   astikit.IntPtr(11),
		STLPublisher:                                        "",
		STLDisplayStandardCode:                              "0",
		STLCountryOfOrigin:                                  "NOR",
		Title:                                               "",
		CreationDate:                                        &creationDate,
		RevisionDate:                                        &revisionDate},
		s.Metadata)

	// No subtitles to write
	w := &bytes.Buffer{}
	err = astisub.Subtitles{}.WriteToSTL(w)
	assert.EqualError(t, err, astisub.ErrNoSubtitlesToWrite.Error())

	// Write
	c, err := ioutil.ReadFile("./testdata/example-opn-out.stl")
	assert.NoError(t, err)
	err = s.WriteToSTL(w)
	assert.NoError(t, err)
	assert.Equal(t, string(c), w.String())
}
