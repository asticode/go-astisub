package astisub

import (
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"sort"

	"github.com/asticode/go-astitools/map"
	"github.com/asticode/go-astitools/string"
	"github.com/pkg/errors"
)

// https://www.w3.org/TR/ttaf1-dfxp/
// http://www.skynav.com:8080/ttv/check
// https://www.speechpad.com/captions/ttml

// TTML languages
const (
	ttmlLanguageEnglish = "en"
	ttmlLanguageFrench  = "fr"
)

// TTML language mapping
var ttmlLanguageMapping = astimap.NewMap(ttmlLanguageEnglish, LanguageEnglish).
	Set(ttmlLanguageFrench, LanguageFrench)

// TTML Clock Time Frames and Offset Time
var (
	ttmlRegexpClockTimeFrames = regexp.MustCompile("\\:[\\d]+$")
	ttmlRegexpOffsetTime      = regexp.MustCompile("^(\\d+)(\\.(\\d+))?(h|m|s|ms|f|t)$")
)

// TTMLIn represents an input TTML that must be unmarshaled
// We split it from the output TTML as we can't add strict namespace without breaking retrocompatibility
type TTMLIn struct {
	Framerate int              `xml:"frameRate,attr"`
	Lang      string           `xml:"lang,attr"`
	Metadata  TTMLInMetadata   `xml:"head>metadata"`
	Regions   []TTMLInRegion   `xml:"head>layout>region"`
	Styles    []TTMLInStyle    `xml:"head>styling>style"`
	Subtitles []TTMLInSubtitle `xml:"body>div>p"`
	XMLName   xml.Name         `xml:"tt"`
}

// metadata returns the Metadata of the TTML
func (t TTMLIn) metadata() *Metadata {
	return &Metadata{
		Framerate:     t.Framerate,
		Language:      ttmlLanguageMapping.B(astistring.ToLength(t.Lang, " ", 2)).(string),
		Title:         t.Metadata.Title,
		TTMLCopyright: t.Metadata.Copyright,
	}
}

// TTMLInMetadata represents an input TTML Metadata
type TTMLInMetadata struct {
	Copyright string `xml:"copyright"`
	Title     string `xml:"title"`
}

// TTMLInStyleAttributes represents input TTML style attributes
type TTMLInStyleAttributes struct {
	BackgroundColor string `xml:"backgroundColor,attr,omitempty"`
	Color           string `xml:"color,attr,omitempty"`
	Direction       string `xml:"direction,attr,omitempty"`
	Display         string `xml:"display,attr,omitempty"`
	DisplayAlign    string `xml:"displayAlign,attr,omitempty"`
	Extent          string `xml:"extent,attr,omitempty"`
	FontFamily      string `xml:"fontFamily,attr,omitempty"`
	FontSize        string `xml:"fontSize,attr,omitempty"`
	FontStyle       string `xml:"fontStyle,attr,omitempty"`
	FontWeight      string `xml:"fontWeight,attr,omitempty"`
	LineHeight      string `xml:"lineHeight,attr,omitempty"`
	Opacity         string `xml:"opacity,attr,omitempty"`
	Origin          string `xml:"origin,attr,omitempty"`
	Overflow        string `xml:"overflow,attr,omitempty"`
	Padding         string `xml:"padding,attr,omitempty"`
	ShowBackground  string `xml:"showBackground,attr,omitempty"`
	TextAlign       string `xml:"textAlign,attr,omitempty"`
	TextDecoration  string `xml:"textDecoration,attr,omitempty"`
	TextOutline     string `xml:"textOutline,attr,omitempty"`
	UnicodeBidi     string `xml:"unicodeBidi,attr,omitempty"`
	Visibility      string `xml:"visibility,attr,omitempty"`
	WrapOption      string `xml:"wrapOption,attr,omitempty"`
	WritingMode     string `xml:"writingMode,attr,omitempty"`
	ZIndex          int    `xml:"zIndex,attr,omitempty"`
}

// StyleAttributes converts TTMLInStyleAttributes into a StyleAttributes
func (s TTMLInStyleAttributes) styleAttributes() (o *StyleAttributes) {
	o = &StyleAttributes{
		TTMLBackgroundColor: s.BackgroundColor,
		TTMLColor:           s.Color,
		TTMLDirection:       s.Direction,
		TTMLDisplay:         s.Display,
		TTMLDisplayAlign:    s.DisplayAlign,
		TTMLExtent:          s.Extent,
		TTMLFontFamily:      s.FontFamily,
		TTMLFontSize:        s.FontSize,
		TTMLFontStyle:       s.FontStyle,
		TTMLFontWeight:      s.FontWeight,
		TTMLLineHeight:      s.LineHeight,
		TTMLOpacity:         s.Opacity,
		TTMLOrigin:          s.Origin,
		TTMLOverflow:        s.Overflow,
		TTMLPadding:         s.Padding,
		TTMLShowBackground:  s.ShowBackground,
		TTMLTextAlign:       s.TextAlign,
		TTMLTextDecoration:  s.TextDecoration,
		TTMLTextOutline:     s.TextOutline,
		TTMLUnicodeBidi:     s.UnicodeBidi,
		TTMLVisibility:      s.Visibility,
		TTMLWrapOption:      s.WrapOption,
		TTMLWritingMode:     s.WritingMode,
		TTMLZIndex:          s.ZIndex,
	}
	o.propagateTTMLAttributes()
	return
}

// TTMLInHeader represents an input TTML header
type TTMLInHeader struct {
	ID    string `xml:"id,attr,omitempty"`
	Style string `xml:"style,attr,omitempty"`
	TTMLInStyleAttributes
}

// TTMLInRegion represents an input TTML region
type TTMLInRegion struct {
	TTMLInHeader
	XMLName xml.Name `xml:"region"`
}

// TTMLInStyle represents an input TTML style
type TTMLInStyle struct {
	TTMLInHeader
	XMLName xml.Name `xml:"style"`
}

// TTMLInSubtitle represents an input TTML subtitle
type TTMLInSubtitle struct {
	Begin  *TTMLInDuration `xml:"begin,attr,omitempty"`
	End    *TTMLInDuration `xml:"end,attr,omitempty"`
	ID     string          `xml:"id,attr,omitempty"`
	Items  string          `xml:",innerxml"` // We must store inner XML here since there's no tag to describe both any tag and chardata
	Region string          `xml:"region,attr,omitempty"`
	Style  string          `xml:"style,attr,omitempty"`
	TTMLInStyleAttributes
}

// TTMLInItems represents input TTML items
type TTMLInItems []TTMLInItem

// UnmarshalXML implements the XML unmarshaler interface
func (i *TTMLInItems) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	// Get next tokens
	var t xml.Token
	for {
		// Get next token
		if t, err = d.Token(); err != nil {
			if err == io.EOF {
				break
			}
			err = errors.Wrap(err, "astisub: getting next token failed")
			return
		}

		// Start element
		if se, ok := t.(xml.StartElement); ok {
			var e = TTMLInItem{}
			if err = d.DecodeElement(&e, &se); err != nil {
				err = errors.Wrap(err, "astisub: decoding xml.StartElement failed")
				return
			}
			*i = append(*i, e)
		} else if b, ok := t.(xml.CharData); ok {
			var str = strings.TrimSpace(string(b))
			if len(str) > 0 {
				*i = append(*i, TTMLInItem{Text: str})
			}
		}
	}
	return nil
}

// TTMLInItem represents an input TTML item
type TTMLInItem struct {
	Style string `xml:"style,attr,omitempty"`
	Text  string `xml:",chardata"`
	TTMLInStyleAttributes
	XMLName xml.Name
}

// TTMLInDuration represents an input TTML duration
type TTMLInDuration struct {
	d                 time.Duration
	frames, framerate int // Framerate is in frame/s
}

// UnmarshalText implements the TextUnmarshaler interface
// Possible formats are:
// - hh:mm:ss.mmm
// - hh:mm:ss:fff (fff being frames)
func (d *TTMLInDuration) UnmarshalText(i []byte) (err error) {
	var text = string(i)
	if matches := ttmlRegexpOffsetTime.FindStringSubmatch(text); matches != nil {
		metric := matches[4]
		value, err := strconv.Atoi(matches[1])

		if err != nil {
			err = errors.Wrapf(err, "astisub: failed to parse value %s", matches[1])
			return err
		}

		d.d = time.Duration(0)

		var (
			nsBase       int64
			fraction     int
			fractionBase float64
		)

		if len(matches[3]) > 0 {
			fraction, err = strconv.Atoi(matches[3])
			fractionBase = math.Pow10(len(matches[3]))

			if err != nil {
				err = errors.Wrapf(err, "astisub: failed to parse fraction %s", matches[3])
				return err
			}
		}

		switch metric {
		case "h":
			nsBase = time.Hour.Nanoseconds()
		case "m":
			nsBase = time.Minute.Nanoseconds()
		case "s":
			nsBase = time.Second.Nanoseconds()
		case "ms":
			nsBase = time.Millisecond.Nanoseconds()
		case "f":
			nsBase = time.Second.Nanoseconds()
			d.frames = value % d.framerate
			value = value / d.framerate
			// TODO: fraction of frames
		case "t":
			// TODO: implement ticks
			return errors.New("astisub: offset time in ticks not implemented")
		}

		d.d += time.Duration(nsBase * int64(value))

		if fractionBase > 0 {
			d.d += time.Duration(nsBase * int64(fraction) / int64(fractionBase))
		}

		return nil

	}
	if indexes := ttmlRegexpClockTimeFrames.FindStringIndex(text); indexes != nil {
		// Parse frames
		var s = text[indexes[0]+1 : indexes[1]]
		if d.frames, err = strconv.Atoi(s); err != nil {
			err = errors.Wrapf(err, "astisub: atoi %s failed", s)
			return
		}

		// Update text
		text = text[:indexes[0]] + ".000"
	}

	d.d, err = parseDuration(text, ".", 3)
	return
}

// duration returns the input TTML Duration's time.Duration
func (d TTMLInDuration) duration() time.Duration {
	if d.framerate > 0 {
		return d.d + time.Duration(float64(d.frames)/float64(d.framerate)*1e9)*time.Nanosecond
	}
	return d.d
}

// ReadFromTTML parses a .ttml content
func ReadFromTTML(i io.Reader) (o *Subtitles, err error) {
	// Init
	o = NewSubtitles()

	// Unmarshal XML
	var ttml TTMLIn
	if err = xml.NewDecoder(i).Decode(&ttml); err != nil {
		err = errors.Wrap(err, "astisub: xml decoding failed")
		return
	}

	// Add metadata
	o.Metadata = ttml.metadata()

	// Loop through styles
	var parentStyles = make(map[string]*Style)
	for _, ts := range ttml.Styles {
		var s = &Style{
			ID:          ts.ID,
			InlineStyle: ts.TTMLInStyleAttributes.styleAttributes(),
		}
		o.Styles[s.ID] = s
		if len(ts.Style) > 0 {
			parentStyles[ts.Style] = s
		}
	}

	// Take care of parent styles
	for id, s := range parentStyles {
		if _, ok := o.Styles[id]; !ok {
			err = fmt.Errorf("astisub: Style %s requested by style %s doesn't exist", id, s.ID)
			return
		}
		s.Style = o.Styles[id]
	}

	// Loop through regions
	for _, tr := range ttml.Regions {
		var r = &Region{
			ID:          tr.ID,
			InlineStyle: tr.TTMLInStyleAttributes.styleAttributes(),
		}
		if len(tr.Style) > 0 {
			if _, ok := o.Styles[tr.Style]; !ok {
				err = fmt.Errorf("astisub: Style %s requested by region %s doesn't exist", tr.Style, r.ID)
				return
			}
			r.Style = o.Styles[tr.Style]
		}
		o.Regions[r.ID] = r
	}

	// Loop through subtitles
	for _, ts := range ttml.Subtitles {
		// Init item
		ts.Begin.framerate = ttml.Framerate
		ts.End.framerate = ttml.Framerate
		var s = &Item{
			EndAt:       ts.End.duration(),
			InlineStyle: ts.TTMLInStyleAttributes.styleAttributes(),
			StartAt:     ts.Begin.duration(),
		}

		// Add region
		if len(ts.Region) > 0 {
			if _, ok := o.Regions[ts.Region]; !ok {
				err = fmt.Errorf("astisub: Region %s requested by subtitle between %s and %s doesn't exist", ts.Region, s.StartAt, s.EndAt)
				return
			}
			s.Region = o.Regions[ts.Region]
		}

		// Add style
		if len(ts.Style) > 0 {
			if _, ok := o.Styles[ts.Style]; !ok {
				err = fmt.Errorf("astisub: Style %s requested by subtitle between %s and %s doesn't exist", ts.Style, s.StartAt, s.EndAt)
				return
			}
			s.Style = o.Styles[ts.Style]
		}

		// Unmarshal items
		var items = TTMLInItems{}
		if err = xml.Unmarshal([]byte("<span>"+ts.Items+"</span>"), &items); err != nil {
			err = errors.Wrap(err, "astisub: unmarshaling items failed")
			return
		}

		// Loop through texts
		var l = &Line{}
		for _, tt := range items {
			// New line specified with the "br" tag
			if strings.ToLower(tt.XMLName.Local) == "br" {
				s.Lines = append(s.Lines, *l)
				l = &Line{}
				continue
			}

			// New line decoded as a line break. This can happen if there's a "br" tag within the text since
			// since the go xml unmarshaler will unmarshal a "br" tag as a line break if the field has the
			// chardata xml tag.
			for idx, li := range strings.Split(tt.Text, "\n") {
				// New line
				if idx > 0 {
					s.Lines = append(s.Lines, *l)
					l = &Line{}
				}

				// Init line item
				var t = LineItem{
					InlineStyle: tt.TTMLInStyleAttributes.styleAttributes(),
					Text:        strings.TrimSpace(li),
				}

				// Add style
				if len(tt.Style) > 0 {
					if _, ok := o.Styles[tt.Style]; !ok {
						err = fmt.Errorf("astisub: Style %s requested by item with text %s doesn't exist", tt.Style, tt.Text)
						return
					}
					t.Style = o.Styles[tt.Style]
				}

				// Append items
				l.Items = append(l.Items, t)
			}

		}
		s.Lines = append(s.Lines, *l)

		// Append subtitle
		o.Items = append(o.Items, s)
	}
	return
}

// TTMLOut represents an output TTML that must be marshaled
// We split it from the input TTML as this time we'll add strict namespaces
type TTMLOut struct {
	Lang            string            `xml:"xml:lang,attr,omitempty"`
	Metadata        *TTMLOutMetadata  `xml:"head>metadata,omitempty"`
	Styles          []TTMLOutStyle    `xml:"head>styling>style,omitempty"` //!\\ Order is important! Keep Styling above Layout
	Regions         []TTMLOutRegion   `xml:"head>layout>region,omitempty"`
	Subtitles       []TTMLOutSubtitle `xml:"body>div>p,omitempty"`
	XMLName         xml.Name          `xml:"http://www.w3.org/ns/ttml tt"`
	XMLNamespaceTTM string            `xml:"xmlns:ttm,attr"`
	XMLNamespaceTTS string            `xml:"xmlns:tts,attr"`
}

// TTMLOutMetadata represents an output TTML Metadata
type TTMLOutMetadata struct {
	Copyright string `xml:"ttm:copyright,omitempty"`
	Title     string `xml:"ttm:title,omitempty"`
}

// TTMLOutStyleAttributes represents output TTML style attributes
type TTMLOutStyleAttributes struct {
	BackgroundColor string `xml:"tts:backgroundColor,attr,omitempty"`
	Color           string `xml:"tts:color,attr,omitempty"`
	Direction       string `xml:"tts:direction,attr,omitempty"`
	Display         string `xml:"tts:display,attr,omitempty"`
	DisplayAlign    string `xml:"tts:displayAlign,attr,omitempty"`
	Extent          string `xml:"tts:extent,attr,omitempty"`
	FontFamily      string `xml:"tts:fontFamily,attr,omitempty"`
	FontSize        string `xml:"tts:fontSize,attr,omitempty"`
	FontStyle       string `xml:"tts:fontStyle,attr,omitempty"`
	FontWeight      string `xml:"tts:fontWeight,attr,omitempty"`
	LineHeight      string `xml:"tts:lineHeight,attr,omitempty"`
	Opacity         string `xml:"tts:opacity,attr,omitempty"`
	Origin          string `xml:"tts:origin,attr,omitempty"`
	Overflow        string `xml:"tts:overflow,attr,omitempty"`
	Padding         string `xml:"tts:padding,attr,omitempty"`
	ShowBackground  string `xml:"tts:showBackground,attr,omitempty"`
	TextAlign       string `xml:"tts:textAlign,attr,omitempty"`
	TextDecoration  string `xml:"tts:textDecoration,attr,omitempty"`
	TextOutline     string `xml:"tts:textOutline,attr,omitempty"`
	UnicodeBidi     string `xml:"tts:unicodeBidi,attr,omitempty"`
	Visibility      string `xml:"tts:visibility,attr,omitempty"`
	WrapOption      string `xml:"tts:wrapOption,attr,omitempty"`
	WritingMode     string `xml:"tts:writingMode,attr,omitempty"`
	ZIndex          int    `xml:"tts:zIndex,attr,omitempty"`
}

// ttmlOutStyleAttributesFromStyleAttributes converts StyleAttributes into a TTMLOutStyleAttributes
func ttmlOutStyleAttributesFromStyleAttributes(s *StyleAttributes) TTMLOutStyleAttributes {
	if s == nil {
		return TTMLOutStyleAttributes{}
	}
	return TTMLOutStyleAttributes{
		BackgroundColor: s.TTMLBackgroundColor,
		Color:           s.TTMLColor,
		Direction:       s.TTMLDirection,
		Display:         s.TTMLDisplay,
		DisplayAlign:    s.TTMLDisplayAlign,
		Extent:          s.TTMLExtent,
		FontFamily:      s.TTMLFontFamily,
		FontSize:        s.TTMLFontSize,
		FontStyle:       s.TTMLFontStyle,
		FontWeight:      s.TTMLFontWeight,
		LineHeight:      s.TTMLLineHeight,
		Opacity:         s.TTMLOpacity,
		Origin:          s.TTMLOrigin,
		Overflow:        s.TTMLOverflow,
		Padding:         s.TTMLPadding,
		ShowBackground:  s.TTMLShowBackground,
		TextAlign:       s.TTMLTextAlign,
		TextDecoration:  s.TTMLTextDecoration,
		TextOutline:     s.TTMLTextOutline,
		UnicodeBidi:     s.TTMLUnicodeBidi,
		Visibility:      s.TTMLVisibility,
		WrapOption:      s.TTMLWrapOption,
		WritingMode:     s.TTMLWritingMode,
		ZIndex:          s.TTMLZIndex,
	}
}

// TTMLOutHeader represents an output TTML header
type TTMLOutHeader struct {
	ID    string `xml:"xml:id,attr,omitempty"`
	Style string `xml:"style,attr,omitempty"`
	TTMLOutStyleAttributes
}

// TTMLOutRegion represents an output TTML region
type TTMLOutRegion struct {
	TTMLOutHeader
	XMLName xml.Name `xml:"region"`
}

// TTMLOutStyle represents an output TTML style
type TTMLOutStyle struct {
	TTMLOutHeader
	XMLName xml.Name `xml:"style"`
}

// TTMLOutSubtitle represents an output TTML subtitle
type TTMLOutSubtitle struct {
	Begin  TTMLOutDuration `xml:"begin,attr"`
	End    TTMLOutDuration `xml:"end,attr"`
	ID     string          `xml:"id,attr,omitempty"`
	Items  []TTMLOutItem
	Region string `xml:"region,attr,omitempty"`
	Style  string `xml:"style,attr,omitempty"`
	TTMLOutStyleAttributes
}

// TTMLOutItem represents an output TTML Item
type TTMLOutItem struct {
	Style string `xml:"style,attr,omitempty"`
	Text  string `xml:",chardata"`
	TTMLOutStyleAttributes
	XMLName xml.Name
}

// TTMLOutDuration represents an output TTML duration
type TTMLOutDuration time.Duration

// MarshalText implements the TextMarshaler interface
func (t TTMLOutDuration) MarshalText() ([]byte, error) {
	return []byte(formatDuration(time.Duration(t), ".", 3)), nil
}

// WriteToTTML writes subtitles in .ttml format
func (s Subtitles) WriteToTTML(o io.Writer) (err error) {
	// Do not write anything if no subtitles
	if len(s.Items) == 0 {
		return ErrNoSubtitlesToWrite
	}

	// Init TTML
	var ttml = TTMLOut{
		XMLNamespaceTTM: "http://www.w3.org/ns/ttml#metadata",
		XMLNamespaceTTS: "http://www.w3.org/ns/ttml#styling",
	}

	// Add metadata
	if s.Metadata != nil {
		ttml.Lang = ttmlLanguageMapping.A(s.Metadata.Language).(string)
		if len(s.Metadata.TTMLCopyright) > 0 || len(s.Metadata.Title) > 0 {
			ttml.Metadata = &TTMLOutMetadata{
				Copyright: s.Metadata.TTMLCopyright,
				Title:     s.Metadata.Title,
			}
		}
	}

	// Add regions
	var k []string
	for _, region := range s.Regions {
		k = append(k, region.ID)
	}
	sort.Strings(k)
	for _, id := range k {
		var ttmlRegion = TTMLOutRegion{TTMLOutHeader: TTMLOutHeader{
			ID: s.Regions[id].ID,
			TTMLOutStyleAttributes: ttmlOutStyleAttributesFromStyleAttributes(s.Regions[id].InlineStyle),
		}}
		if s.Regions[id].Style != nil {
			ttmlRegion.Style = s.Regions[id].Style.ID
		}
		ttml.Regions = append(ttml.Regions, ttmlRegion)
	}

	// Add styles
	k = []string{}
	for _, style := range s.Styles {
		k = append(k, style.ID)
	}
	sort.Strings(k)
	for _, id := range k {
		var ttmlStyle = TTMLOutStyle{TTMLOutHeader: TTMLOutHeader{
			ID: s.Styles[id].ID,
			TTMLOutStyleAttributes: ttmlOutStyleAttributesFromStyleAttributes(s.Styles[id].InlineStyle),
		}}
		if s.Styles[id].Style != nil {
			ttmlStyle.Style = s.Styles[id].Style.ID
		}
		ttml.Styles = append(ttml.Styles, ttmlStyle)
	}

	// Add items
	for _, item := range s.Items {
		// Init subtitle
		var ttmlSubtitle = TTMLOutSubtitle{
			Begin: TTMLOutDuration(item.StartAt),
			End:   TTMLOutDuration(item.EndAt),
			TTMLOutStyleAttributes: ttmlOutStyleAttributesFromStyleAttributes(item.InlineStyle),
		}

		// Add region
		if item.Region != nil {
			ttmlSubtitle.Region = item.Region.ID
		}

		// Add style
		if item.Style != nil {
			ttmlSubtitle.Style = item.Style.ID
		}

		// Add lines
		for _, line := range item.Lines {
			// Loop through line items
			for _, lineItem := range line.Items {
				// Init ttml item
				var ttmlItem = TTMLOutItem{
					Text: lineItem.Text,
					TTMLOutStyleAttributes: ttmlOutStyleAttributesFromStyleAttributes(lineItem.InlineStyle),
					XMLName:                xml.Name{Local: "span"},
				}

				// Add style
				if lineItem.Style != nil {
					ttmlItem.Style = lineItem.Style.ID
				}

				// Add ttml item
				ttmlSubtitle.Items = append(ttmlSubtitle.Items, ttmlItem)
			}

			// Add line break
			ttmlSubtitle.Items = append(ttmlSubtitle.Items, TTMLOutItem{XMLName: xml.Name{Local: "br"}})
		}

		// Remove last line break
		if len(ttmlSubtitle.Items) > 0 {
			ttmlSubtitle.Items = ttmlSubtitle.Items[:len(ttmlSubtitle.Items)-1]
		}

		// Append subtitle
		ttml.Subtitles = append(ttml.Subtitles, ttmlSubtitle)
	}

	// Marshal XML
	var e = xml.NewEncoder(o)
	e.Indent("", "    ")
	if err = e.Encode(ttml); err != nil {
		err = errors.Wrap(err, "astisub: xml encoding failed")
		return
	}
	return
}
