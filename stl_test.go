package astisub_test

import (
	"bytes"
	"io/ioutil"
	"os"
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
		Framerate:       25,
		Language:        astisub.LanguageFrench,
		STLCreationDate: &creationDate,
		STLMaximumNumberOfDisplayableCharactersInAnyTextRow: astikit.IntPtr(40),
		STLMaximumNumberOfDisplayableRows:                   astikit.IntPtr(23),
		STLPublisher:                                        "Copyright test",
		STLDisplayStandardCode:                              "1",
		STLRevisionDate:                                     &revisionDate,
		STLSubtitleListReferenceCode:                        "12345678",
		STLCountryOfOrigin:                                  "FRA",
		Title:                                               "Title test"},
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
	// Metadata
	assert.Equal(t, &astisub.Metadata{
		Framerate:              25,
		Language:               astisub.LanguageEnglish,
		STLCountryOfOrigin:     "NOR",
		STLCreationDate:        &creationDate,
		STLDisplayStandardCode: "0",
		STLMaximumNumberOfDisplayableCharactersInAnyTextRow: astikit.IntPtr(38),
		STLMaximumNumberOfDisplayableRows:                   astikit.IntPtr(11),
		STLPublisher:                                        "",
		STLRevisionDate:                                     &revisionDate,
		Title:                                               ""},
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

func TestIgnoreTimecodeStartOfProgramme(t *testing.T) {
	opts := astisub.STLOptions{IgnoreTimecodeStartOfProgramme: true}
	r, err := os.Open("./testdata/example-in-nonzero-offset.stl")
	assert.NoError(t, err)
	defer r.Close()

	s, err := astisub.ReadFromSTL(r, opts)
	assert.NoError(t, err)
	firstStart := 99 * time.Second
	assert.Equal(t, firstStart, s.Items[0].StartAt, "first start at 0")
}
