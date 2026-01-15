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
		STLRevisionNumber:                                   1,
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

func TestTTMLToSTLGSIBlock(t *testing.T) {
	// Test that TTML to STL conversion includes correct disk format code and display standard code
	// This verifies the fix for the bug where "STL25.01" was missing from the GSI block
	// when converting from TTML files that have no framerate metadata

	// Open TTML file
	s, err := astisub.OpenFile("./testdata/example-in.ttml")
	assert.NoError(t, err)
	assert.NotNil(t, s)

	// TTML files typically don't have STL-specific metadata
	assert.Empty(t, s.Metadata.STLDisplayStandardCode, "TTML should not have STL-specific display standard code")

	// Write to STL
	w := &bytes.Buffer{}
	err = s.WriteToSTL(w)
	assert.NoError(t, err)

	stlData := w.Bytes()
	assert.True(t, len(stlData) >= 1024, "STL file should have at least GSI block (1024 bytes)")

	// Check GSI block header
	// Bytes 0-2: Code page number (should be "850")
	codePageNumber := string(stlData[0:3])
	assert.Equal(t, "850", codePageNumber, "Code page number should be '850'")

	// Bytes 3-10: Disk format code (should be "STL25.01" for 25fps, not spaces)
	diskFormatCode := string(stlData[3:11])
	assert.NotEqual(t, "        ", diskFormatCode,
		"Disk format code should not be empty spaces (this was the bug)")
	assert.True(t, diskFormatCode == "STL25.01" || diskFormatCode == "STL30.01",
		"Disk format code should be 'STL25.01' or 'STL30.01', got '%s'", diskFormatCode)

	// Byte 11: Display standard code (should be "1" for Level 1 Teletext by default)
	displayStandardCode := string(stlData[11:12])
	assert.True(t, displayStandardCode == "0" || displayStandardCode == "1" || displayStandardCode == "2",
		"Display standard code should be '0', '1', or '2', got '%s'", displayStandardCode)
}
