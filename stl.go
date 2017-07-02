package astisub

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/asticode/go-astitools/byte"
	"github.com/asticode/go-astitools/map"
	"github.com/pkg/errors"
	"golang.org/x/text/unicode/norm"
)

// https://tech.ebu.ch/docs/tech/tech3264.pdf
// https://github.com/yanncoupin/stl2srt/blob/master/to_srt.py

// STL block sizes
const (
	stlBlockSizeGSI = 1024
	stlBlockSizeTTI = 128
)

// STL character code table number
const (
	stlCharacterCodeTableNumberLatin         = "00"
	stlCharacterCodeTableNumberLatinCyrillic = "01"
	stlCharacterCodeTableNumberLatinArabic   = "02"
	stlCharacterCodeTableNumberLatinGreek    = "03"
	stlCharacterCodeTableNumberLatinHebrew   = "04"
)

// STL code page numbers
const (
	stlCodePageNumberCanadaFrench = "863"
	stlCodePageNumberMultilingual = "850"
	stlCodePageNumberNordic       = "865"
	stlCodePageNumberPortugal     = "860"
	stlCodePageNumberUnitedStates = "437"
)

// STL comment flag
const (
	stlCommentFlagTextContainsSubtitleData                       = '\x00'
	stlCommentFlagTextContainsCommentsNotIntendedForTransmission = '\x01'
)

// STL country codes
const (
	stlCountryCodeFrance = "FRA"
)

// STL cumulative status
const (
	stlCumulativeStatusFirstSubtitleOfACumulativeSet        = '\x01'
	stlCumulativeStatusIntermediateSubtitleOfACumulativeSet = '\x02'
	stlCumulativeStatusLastSubtitleOfACumulativeSet         = '\x03'
	stlCumulativeStatusSubtitleNotPartOfACumulativeSet      = '\x00'
)

// STL display standard code
const (
	stlDisplayStandardCodeOpenSubtitling = "0"
	stlDisplayStandardCodeLevel1Teletext = "1"
	stlDisplayStandardCodeLevel2Teletext = "2"
)

// STL framerate mapping
var stlFramerateMapping = astimap.NewMap("STL25.01", 25).
	Set("STL25.01", 25).
	Set("STL30.01", 30)

// STL justification code
const (
	stlJustificationCodeCentredText           = '\x02'
	stlJustificationCodeLeftJustifiedText     = '\x01'
	stlJustificationCodeRightJustifiedText    = '\x03'
	stlJustificationCodeUnchangedPresentation = '\x00'
)

// STL language codes
const (
	stlLanguageCodeEnglish = "09"
	stlLanguageCodeFrench  = "0F"
)

// STL language mapping
var stlLanguageMapping = astimap.NewMap(stlLanguageCodeEnglish, LanguageEnglish).
	Set(stlLanguageCodeFrench, LanguageFrench)

// STL timecode status
const (
	stlTimecodeStatusNotIntendedForUse = "0"
	stlTimecodeStatusIntendedForUse    = "1"
)

// STL unicode diacritic
var stlUnicodeDiacritic = astimap.NewMap(byte('\x00'), "\x00").
	Set(byte('\xc1'), "\u0300"). // Grave accent
	Set(byte('\xc2'), "\u0301"). // Acute accent
	Set(byte('\xc3'), "\u0302"). // Circumflex
	Set(byte('\xc4'), "\u0303"). // Tilde
	Set(byte('\xc5'), "\u0304"). // Macron
	Set(byte('\xc6'), "\u0306"). // Breve
	Set(byte('\xc7'), "\u0307"). // Dot
	Set(byte('\xc8'), "\u0308"). // Umlaut
	Set(byte('\xca'), "\u030a"). // Ring
	Set(byte('\xcb'), "\u0327"). // Cedilla
	Set(byte('\xcd'), "\u030B"). // Double acute accent
	Set(byte('\xce'), "\u0328"). // Ogonek
	Set(byte('\xcf'), "\u030c")  // Caron

// STL unicode mapping
var stlUnicodeMapping = astimap.NewMap(byte('\x00'), "\x00").
	Set(byte('\x8a'), "\u000a"). // Line break
	Set(byte('\xa8'), "\u00a4"). // ¤
	Set(byte('\xa9'), "\u2018"). // ‘
	Set(byte('\xaa'), "\u201C"). // “
	Set(byte('\xab'), "\u00AB"). // «
	Set(byte('\xac'), "\u2190"). // ←
	Set(byte('\xad'), "\u2191"). // ↑
	Set(byte('\xae'), "\u2192"). // →
	Set(byte('\xaf'), "\u2193"). // ↓
	Set(byte('\xb4'), "\u00D7"). // ×
	Set(byte('\xb8'), "\u00F7"). // ÷
	Set(byte('\xb9'), "\u2019"). // ’
	Set(byte('\xba'), "\u201D"). // ”
	Set(byte('\xbc'), "\u00BC"). // ¼
	Set(byte('\xbd'), "\u00BD"). // ½
	Set(byte('\xbe'), "\u00BE"). // ¾
	Set(byte('\xbf'), "\u00BF"). // ¿
	Set(byte('\xd0'), "\u2015"). // ―
	Set(byte('\xd1'), "\u00B9"). // ¹
	Set(byte('\xd2'), "\u00AE"). // ®
	Set(byte('\xd3'), "\u00A9"). // ©
	Set(byte('\xd4'), "\u2122"). // ™
	Set(byte('\xd5'), "\u266A"). // ♪
	Set(byte('\xd6'), "\u00AC"). // ¬
	Set(byte('\xd7'), "\u00A6"). // ¦
	Set(byte('\xdc'), "\u215B"). // ⅛
	Set(byte('\xdd'), "\u215C"). // ⅜
	Set(byte('\xde'), "\u215D"). // ⅝
	Set(byte('\xdf'), "\u215E"). // ⅞
	Set(byte('\xe0'), "\u2126"). // Ohm Ω
	Set(byte('\xe1'), "\u00C6"). // Æ
	Set(byte('\xe2'), "\u0110"). // Đ
	Set(byte('\xe3'), "\u00AA"). // ª
	Set(byte('\xe4'), "\u0126"). // Ħ
	Set(byte('\xe6'), "\u0132"). // Ĳ
	Set(byte('\xe7'), "\u013F"). // Ŀ
	Set(byte('\xe8'), "\u0141"). // Ł
	Set(byte('\xe9'), "\u00D8"). // Ø
	Set(byte('\xea'), "\u0152"). // Œ
	Set(byte('\xeb'), "\u00BA"). // º
	Set(byte('\xec'), "\u00DE"). // Þ
	Set(byte('\xed'), "\u0166"). // Ŧ
	Set(byte('\xee'), "\u014A"). // Ŋ
	Set(byte('\xef'), "\u0149"). // ŉ
	Set(byte('\xf0'), "\u0138"). // ĸ
	Set(byte('\xf1'), "\u00E6"). // æ
	Set(byte('\xf2'), "\u0111"). // đ
	Set(byte('\xf3'), "\u00F0"). // ð
	Set(byte('\xf4'), "\u0127"). // ħ
	Set(byte('\xf5'), "\u0131"). // ı
	Set(byte('\xf6'), "\u0133"). // ĳ
	Set(byte('\xf7'), "\u0140"). // ŀ
	Set(byte('\xf8'), "\u0142"). // ł
	Set(byte('\xf9'), "\u00F8"). // ø
	Set(byte('\xfa'), "\u0153"). // œ
	Set(byte('\xfb'), "\u00DF"). // ß
	Set(byte('\xfc'), "\u00FE"). // þ
	Set(byte('\xfd'), "\u0167"). // ŧ
	Set(byte('\xfe'), "\u014B"). // ŋ
	Set(byte('\xff'), "\u00AD")  // Soft hyphen

// ReadFromSTL parses an .stl content
func ReadFromSTL(i io.Reader) (o *Subtitles, err error) {
	// Init
	o = &Subtitles{}

	// Read GSI block
	var b []byte
	if b, err = readNBytes(i, stlBlockSizeGSI); err != nil {
		return
	}

	// Parse GSI block
	var g *gsiBlock
	if g, err = parseGSIBlock(b); err != nil {
		err = errors.Wrap(err, "building gsi block failed")
		return
	}

	// Update metadata
	o.Metadata = &Metadata{
		Copyright: g.publisher,
		Framerate: g.framerate,
		Language:  stlLanguageMapping.B(g.languageCode).(string),
		Title:     g.originalProgramTitle,
	}

	// Parse Text and Timing Information (TTI) blocks.
	for {
		// Read TTI block
		if b, err = readNBytes(i, stlBlockSizeTTI); err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			return
		}

		// Init item
		var t = parseTTIBlock(b, g.framerate)
		var i = &Item{
			EndAt:   t.timecodeOut - g.timecodeStartOfProgramme,
			StartAt: t.timecodeIn - g.timecodeStartOfProgramme,
		}

		// Add lines
		for _, text := range strings.Split(t.text, "\n") {
			text = strings.TrimSpace(text)
			if len(text) == 0 {
				continue
			}
			i.Lines = append(i.Lines, []LineItem{{Text: text}})
		}

		// Append item
		o.Items = append(o.Items, i)
	}
	return
}

// readNBytes reads n bytes
func readNBytes(i io.Reader, c int) (o []byte, err error) {
	o = make([]byte, c)
	var n int
	if n, err = i.Read(o); err != nil || n != len(o) {
		if err != nil {
			if err == io.EOF {
				return
			}
			err = errors.Wrapf(err, "reading %d bytes failed", c)
			return
		} else {
			err = fmt.Errorf("Read %d bytes, should have read %d", n, c)
			return
		}
	}
	return
}

// gsiBlock represents a GSI block
type gsiBlock struct {
	characterCodeTableNumber                         string
	codePageNumber                                   string
	countryOfOrigin                                  string
	creationDate                                     time.Time
	diskSequenceNumber                               int
	displayStandardCode                              string
	editorContactDetails                             string
	editorName                                       string
	framerate                                        int
	languageCode                                     string
	maximumNumberOfDisplayableCharactersInAnyTextRow int
	maximumNumberOfDisplayableRows                   int
	originalEpisodeTitle                             string
	originalProgramTitle                             string
	publisher                                        string
	revisionDate                                     time.Time
	revisionNumber                                   int
	subtitleListReferenceCode                        string
	timecodeFirstInCue                               time.Duration
	timecodeStartOfProgramme                         time.Duration
	timecodeStatus                                   string
	totalNumberOfDisks                               int
	totalNumberOfSubtitleGroups                      int
	totalNumberOfSubtitles                           int
	totalNumberOfTTIBlocks                           int
	translatedEpisodeTitle                           string
	translatedProgramTitle                           string
	translatorContactDetails                         string
	translatorName                                   string
	userDefinedArea                                  string
}

// newGSIBlock builds the subtitles GSI block
func newGSIBlock(s Subtitles) (g *gsiBlock) {
	// Init
	g = &gsiBlock{
		characterCodeTableNumber: stlCharacterCodeTableNumberLatin,
		codePageNumber:           stlCodePageNumberMultilingual,
		countryOfOrigin:          stlCountryCodeFrance,
		creationDate:             time.Now(),
		diskSequenceNumber:       1,
		displayStandardCode:      stlDisplayStandardCodeLevel1Teletext,
		framerate:                25,
		languageCode:             stlLanguageCodeFrench,
		maximumNumberOfDisplayableCharactersInAnyTextRow: 40,
		maximumNumberOfDisplayableRows:                   23,
		subtitleListReferenceCode:                        "12345678",
		timecodeStatus:                                   stlTimecodeStatusIntendedForUse,
		totalNumberOfDisks:                               1,
		totalNumberOfSubtitleGroups:                      1,
		totalNumberOfSubtitles:                           len(s.Items),
		totalNumberOfTTIBlocks:                           len(s.Items),
	}

	// Add metadata
	if s.Metadata != nil {
		g.framerate = s.Metadata.Framerate
		g.languageCode = stlLanguageMapping.A(s.Metadata.Language).(string)
		g.originalProgramTitle = s.Metadata.Title
		g.publisher = s.Metadata.Copyright
	}

	// Timecode first in cue
	if len(s.Items) > 0 {
		g.timecodeFirstInCue = s.Items[0].StartAt
	}
	return
}

// parseGSIBlock parses a GSI block
func parseGSIBlock(b []byte) (g *gsiBlock, err error) {
	// Init
	g = &gsiBlock{
		characterCodeTableNumber:  string(bytes.TrimSpace(b[12:14])),
		countryOfOrigin:           string(bytes.TrimSpace(b[274:277])),
		codePageNumber:            string(bytes.TrimSpace(b[0:3])),
		displayStandardCode:       string(bytes.TrimSpace([]byte{b[11]})),
		editorName:                string(bytes.TrimSpace(b[309:341])),
		editorContactDetails:      string(bytes.TrimSpace(b[341:373])),
		framerate:                 stlFramerateMapping.B(string(b[3:11])).(int),
		languageCode:              string(bytes.TrimSpace(b[14:16])),
		originalEpisodeTitle:      string(bytes.TrimSpace(b[48:80])),
		originalProgramTitle:      string(bytes.TrimSpace(b[16:48])),
		publisher:                 string(bytes.TrimSpace(b[277:309])),
		subtitleListReferenceCode: string(bytes.TrimSpace(b[208:224])),
		timecodeStatus:            string(bytes.TrimSpace([]byte{b[255]})),
		translatedEpisodeTitle:    string(bytes.TrimSpace(b[80:112])),
		translatedProgramTitle:    string(bytes.TrimSpace(b[112:144])),
		translatorContactDetails:  string(bytes.TrimSpace(b[176:208])),
		translatorName:            string(bytes.TrimSpace(b[144:176])),
		userDefinedArea:           string(bytes.TrimSpace(b[448:])),
	}

	// Creation date
	var cd = string(b[224:230])
	if g.creationDate, err = time.Parse("060102", cd); err != nil {
		err = errors.Wrapf(err, "parsing date %s failed", cd)
		return
	}

	// Revision date
	var rd = string(b[230:236])
	if g.revisionDate, err = time.Parse("060102", rd); err != nil {
		err = errors.Wrapf(err, "parsing date %s failed", rd)
		return
	}

	// Revision number
	var rn = string(b[236:238])
	if g.revisionNumber, err = strconv.Atoi(strings.TrimSpace(rn)); err != nil {
		err = errors.Wrapf(err, "atoi of %s failed", rn)
		return
	}

	// Total number of TTI blocks
	var tnb = string(b[238:243])
	if g.totalNumberOfTTIBlocks, err = strconv.Atoi(strings.TrimSpace(tnb)); err != nil {
		err = errors.Wrapf(err, "atoi of %s failed", tnb)
		return
	}

	// Total number of subtitles
	var tns = string(b[243:248])
	if g.totalNumberOfSubtitles, err = strconv.Atoi(strings.TrimSpace(tns)); err != nil {
		err = errors.Wrapf(err, "atoi of %s failed", tns)
		return
	}

	// Total number of subtitle groups
	var tng = string(b[248:251])
	if g.totalNumberOfSubtitleGroups, err = strconv.Atoi(strings.TrimSpace(tng)); err != nil {
		err = errors.Wrapf(err, "atoi of %s failed", tng)
		return
	}

	// Maximum number of displayable characters in any text row
	var mnc = string(b[251:253])
	if g.maximumNumberOfDisplayableCharactersInAnyTextRow, err = strconv.Atoi(strings.TrimSpace(mnc)); err != nil {
		err = errors.Wrapf(err, "atoi of %s failed", mnc)
		return
	}

	// Maximum number of displayable rows
	var mnr = string(b[253:255])
	if g.maximumNumberOfDisplayableRows, err = strconv.Atoi(strings.TrimSpace(mnr)); err != nil {
		err = errors.Wrapf(err, "atoi of %s failed", mnr)
		return
	}

	// Timecode start of programme
	var tcp = string(b[256:264])
	if g.timecodeStartOfProgramme, err = parseDurationSTL(tcp, g.framerate); err != nil {
		err = errors.Wrapf(err, "parsing of stl duration %s failed", tcp)
		return
	}

	// Timecode first in cue
	var tcf = string(b[264:272])
	if g.timecodeFirstInCue, err = parseDurationSTL(tcf, g.framerate); err != nil {
		err = errors.Wrapf(err, "parsing of stl duration %s failed", tcf)
		return
	}

	// Total number of disks
	var tnd = string(b[272])
	if g.totalNumberOfDisks, err = strconv.Atoi(strings.TrimSpace(tnd)); err != nil {
		err = errors.Wrapf(err, "atoi of %s failed", tnd)
		return
	}

	// Disk sequence number
	var dsn = string(b[273])
	if g.diskSequenceNumber, err = strconv.Atoi(strings.TrimSpace(dsn)); err != nil {
		err = errors.Wrapf(err, "atoi of %s failed", dsn)
		return
	}
	return
}

// bytes transforms the GSI block into []byte
func (b gsiBlock) bytes() (o []byte) {
	o = append(o, astibyte.ToLength([]byte(b.codePageNumber), ' ', 3)...)                                                                           // Code page number
	o = append(o, astibyte.ToLength([]byte(stlFramerateMapping.A(b.framerate).(string)), ' ', 8)...)                                                // Disk format code
	o = append(o, astibyte.ToLength([]byte(b.displayStandardCode), ' ', 1)...)                                                                      // Display standard code
	o = append(o, astibyte.ToLength([]byte(b.characterCodeTableNumber), ' ', 2)...)                                                                 // Character code table number
	o = append(o, astibyte.ToLength([]byte(b.languageCode), ' ', 2)...)                                                                             // Language code
	o = append(o, astibyte.ToLength([]byte(b.originalProgramTitle), ' ', 32)...)                                                                    // Original program title
	o = append(o, astibyte.ToLength([]byte(b.originalEpisodeTitle), ' ', 32)...)                                                                    // Original episode title
	o = append(o, astibyte.ToLength([]byte(b.translatedProgramTitle), ' ', 32)...)                                                                  // Translated program title
	o = append(o, astibyte.ToLength([]byte(b.translatedEpisodeTitle), ' ', 32)...)                                                                  // Translated episode title
	o = append(o, astibyte.ToLength([]byte(b.translatorName), ' ', 32)...)                                                                          // Translator's name
	o = append(o, astibyte.ToLength([]byte(b.translatorContactDetails), ' ', 32)...)                                                                // Translator's contact details
	o = append(o, astibyte.ToLength([]byte(b.subtitleListReferenceCode), ' ', 16)...)                                                               // Subtitle list reference code
	o = append(o, astibyte.ToLength([]byte(b.creationDate.Format("060102")), ' ', 6)...)                                                            // Creation date
	o = append(o, astibyte.ToLength([]byte(b.revisionDate.Format("060102")), ' ', 6)...)                                                            // Revision date
	o = append(o, astibyte.ToLength(astibyte.PadLeft([]byte(strconv.Itoa(b.revisionNumber)), '0', 2), '0', 2)...)                                   // Revision number
	o = append(o, astibyte.ToLength(astibyte.PadLeft([]byte(strconv.Itoa(b.totalNumberOfTTIBlocks)), '0', 5), '0', 5)...)                           // Total number of TTI blocks
	o = append(o, astibyte.ToLength(astibyte.PadLeft([]byte(strconv.Itoa(b.totalNumberOfSubtitles)), '0', 5), '0', 5)...)                           // Total number of subtitles
	o = append(o, astibyte.ToLength(astibyte.PadLeft([]byte(strconv.Itoa(b.totalNumberOfSubtitleGroups)), '0', 3), '0', 3)...)                      // Total number of subtitle groups
	o = append(o, astibyte.ToLength(astibyte.PadLeft([]byte(strconv.Itoa(b.maximumNumberOfDisplayableCharactersInAnyTextRow)), '0', 2), '0', 2)...) // Maximum number of displayable characters in any text row
	o = append(o, astibyte.ToLength(astibyte.PadLeft([]byte(strconv.Itoa(b.maximumNumberOfDisplayableRows)), '0', 2), '0', 2)...)                   // Maximum number of displayable rows
	o = append(o, astibyte.ToLength([]byte(b.timecodeStatus), ' ', 1)...)                                                                           // Timecode status
	o = append(o, astibyte.ToLength([]byte(formatDurationSTL(b.timecodeStartOfProgramme, b.framerate)), ' ', 8)...)                                 // Timecode start of a programme
	o = append(o, astibyte.ToLength([]byte(formatDurationSTL(b.timecodeFirstInCue, b.framerate)), ' ', 8)...)                                       // Timecode first in cue
	o = append(o, astibyte.ToLength([]byte(strconv.Itoa(b.totalNumberOfDisks)), ' ', 1)...)                                                         // Total number of disks
	o = append(o, astibyte.ToLength([]byte(strconv.Itoa(b.diskSequenceNumber)), ' ', 1)...)                                                         // Disk sequence number
	o = append(o, astibyte.ToLength([]byte(b.countryOfOrigin), ' ', 3)...)                                                                          // Country of origin
	o = append(o, astibyte.ToLength([]byte(b.publisher), ' ', 32)...)                                                                               // Publisher
	o = append(o, astibyte.ToLength([]byte(b.editorName), ' ', 32)...)                                                                              // Editor's name
	o = append(o, astibyte.ToLength([]byte(b.editorContactDetails), ' ', 32)...)                                                                    // Editor's contact details
	o = append(o, astibyte.ToLength([]byte{}, ' ', 75+576)...)                                                                                      // Spare bytes + user defined area                                                                                           //                                                                                                                      // Editor's contact details
	return
}

// parseDurationSTL parses a STL duration
func parseDurationSTL(i string, framerate int) (d time.Duration, err error) {
	// Parse hours
	var hours, hoursString = 0, i[0:2]
	if hours, err = strconv.Atoi(hoursString); err != nil {
		err = errors.Wrapf(err, "atoi of %s failed", hoursString)
		return
	}

	// Parse minutes
	var minutes, minutesString = 0, i[2:4]
	if minutes, err = strconv.Atoi(minutesString); err != nil {
		err = errors.Wrapf(err, "atoi of %s failed", minutesString)
		return
	}

	// Parse seconds
	var seconds, secondsString = 0, i[4:6]
	if seconds, err = strconv.Atoi(secondsString); err != nil {
		err = errors.Wrapf(err, "atoi of %s failed", secondsString)
		return
	}

	// Parse frames
	var frames, framesString = 0, i[6:8]
	if frames, err = strconv.Atoi(framesString); err != nil {
		err = errors.Wrapf(err, "atoi of %s failed", framesString)
		return
	}

	// Set duration
	d = time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second + time.Duration(1e9*frames/framerate)*time.Nanosecond
	return
}

// formatDurationSTL formats a STL duration
func formatDurationSTL(d time.Duration, framerate int) (o string) {
	// Add hours
	if d.Hours() < 10 {
		o += "0"
	}
	var delta = int(math.Floor(d.Hours()))
	o += strconv.Itoa(delta)
	d -= time.Duration(delta) * time.Hour

	// Add minutes
	if d.Minutes() < 10 {
		o += "0"
	}
	delta = int(math.Floor(d.Minutes()))
	o += strconv.Itoa(delta)
	d -= time.Duration(delta) * time.Minute

	// Add seconds
	if d.Seconds() < 10 {
		o += "0"
	}
	delta = int(math.Floor(d.Seconds()))
	o += strconv.Itoa(delta)
	d -= time.Duration(delta) * time.Second

	// Add frames
	var frames = int(int(d.Nanoseconds()) * framerate / 1e9)
	if frames < 10 {
		o += "0"
	}
	o += strconv.Itoa(frames)
	return
}

// ttiBlock represents a TTI block
type ttiBlock struct {
	commentFlag          byte
	cumulativeStatus     byte
	extensionBlockNumber int
	justificationCode    byte
	subtitleGroupNumber  int
	subtitleNumber       int
	text                 string
	timecodeIn           time.Duration
	timecodeOut          time.Duration
	verticalPosition     int
}

// newTTIBlock builds an item TTI block
func newTTIBlock(i *Item, idx int) (t *ttiBlock) {
	// Init
	t = &ttiBlock{
		commentFlag:          stlCommentFlagTextContainsSubtitleData,
		cumulativeStatus:     stlCumulativeStatusSubtitleNotPartOfACumulativeSet,
		extensionBlockNumber: 255,
		justificationCode:    stlJustificationCodeLeftJustifiedText,
		subtitleGroupNumber:  0,
		subtitleNumber:       idx,
		timecodeIn:           i.StartAt,
		timecodeOut:          i.EndAt,
		verticalPosition:     20,
	}

	// Add text
	var lines []string
	for _, l := range i.Lines {
		lines = append(lines, l.String())
	}
	t.text = strings.Join(lines, "\n")
	return
}

// parseTTIBlock parses a TTI block
func parseTTIBlock(p []byte, framerate int) *ttiBlock {
	return &ttiBlock{
		commentFlag:          p[15],
		cumulativeStatus:     p[4],
		extensionBlockNumber: int(uint8(p[3])),
		justificationCode:    p[14],
		subtitleGroupNumber:  int(uint8(p[0])),
		subtitleNumber:       int(binary.LittleEndian.Uint16(p[1:3])),
		text:                 strings.TrimSpace(decodeTextSTL(p[16:128])),
		timecodeIn:           parseDurationSTLBytes(p[5:9], framerate),
		timecodeOut:          parseDurationSTLBytes(p[9:13], framerate),
		verticalPosition:     int(uint8(p[13])),
	}
}

// bytes transforms the TTI block into []byte
func (t *ttiBlock) bytes(g *gsiBlock) (o []byte) {
	o = append(o, byte(uint8(t.subtitleGroupNumber))) // Subtitle group number
	var b = make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(t.subtitleNumber))
	o = append(o, b...)                                                     // Subtitle number
	o = append(o, byte(uint8(t.extensionBlockNumber)))                      // Extension block number
	o = append(o, t.cumulativeStatus)                                       // Cumulative status
	o = append(o, formatDurationSTLBytes(t.timecodeIn, g.framerate)...)     // Timecode in
	o = append(o, formatDurationSTLBytes(t.timecodeOut, g.framerate)...)    // Timecode out
	o = append(o, byte(uint8(t.verticalPosition)))                          // Vertical position
	o = append(o, t.justificationCode)                                      // Justification code
	o = append(o, t.commentFlag)                                            // Comment flag
	o = append(o, astibyte.ToLength(encodeTextSTL(t.text), '\x8f', 112)...) // Text field
	return
}

// formatDurationSTLBytes formats a STL duration in bytes
func formatDurationSTLBytes(d time.Duration, framerate int) (o []byte) {
	// Add hours
	var hours = int(math.Floor(d.Hours()))
	o = append(o, byte(uint8(hours)))
	d -= time.Duration(hours) * time.Hour

	// Add minutes
	var minutes = int(math.Floor(d.Minutes()))
	o = append(o, byte(uint8(minutes)))
	d -= time.Duration(minutes) * time.Minute

	// Add seconds
	var seconds = int(math.Floor(d.Seconds()))
	o = append(o, byte(uint8(seconds)))
	d -= time.Duration(seconds) * time.Second

	// Add frames
	var frames = int(int(d.Nanoseconds()) * framerate / 1e9)
	o = append(o, byte(uint8(frames)))
	return
}

// parseDurationSTLBytes parses a STL duration in bytes
func parseDurationSTLBytes(b []byte, framerate int) time.Duration {
	return time.Duration(uint8(b[0]))*time.Hour + time.Duration(uint8(b[1]))*time.Minute + time.Duration(uint8(b[2]))*time.Second + time.Duration(1e9*int(uint8(b[3]))/framerate)*time.Nanosecond
}

// encodeTextSTL encodes the STL text
func encodeTextSTL(i string) (o []byte) {
	i = string(norm.NFD.Bytes([]byte(i)))
	for _, c := range i {
		if stlUnicodeMapping.InB(string(c)) {
			o = append(o, stlUnicodeMapping.A(string(c)).(byte))
		} else if stlUnicodeDiacritic.InB(string(c)) {
			o = append(o[:len(o)-1], stlUnicodeDiacritic.A(string(c)).(byte), o[len(o)-1])
		} else {
			o = append(o, byte(c))
		}
	}
	return
}

// decodeTextSTL decodes the STL text
func decodeTextSTL(i []byte) (o string) {
	var state = ""
	for _, c := range i {
		if len(state) == 0 && stlUnicodeMapping.InA(c) {
			o += stlUnicodeMapping.B(c).(string)
		} else if len(state) == 0 && stlUnicodeDiacritic.InA(c) {
			state = stlUnicodeDiacritic.B(c).(string)
		} else if len(state) > 0 {
			o += string(norm.NFC.Bytes([]byte(string(c) + state)))
			state = ""
		} else if c != '\x8f' {
			o += string(c)
		}
	}
	return
}

// WriteToSTL writes subtitles in .stl format
func (s Subtitles) WriteToSTL(o io.Writer) (err error) {
	// Do not write anything if no subtitles
	if len(s.Items) == 0 {
		err = ErrNoSubtitlesToWrite
		return
	}

	// Write GSI block
	var g = newGSIBlock(s)
	if _, err = o.Write(g.bytes()); err != nil {
		err = errors.Wrap(err, "writing gsi block failed")
		return
	}

	// Loop through items
	for idx, item := range s.Items {
		// Write tti block
		if _, err = o.Write(newTTIBlock(item, idx+1).bytes(g)); err != nil {
			err = errors.Wrapf(err, "writing tti block #%d failed", idx+1)
			return
		}
	}
	return
}
