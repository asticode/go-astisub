package astisub

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Constants
const (
	srtTimeBoundariesSeparator = " --> "
)

// Vars
var (
	bytesSRTTimeBoundariesSeparator = []byte(srtTimeBoundariesSeparator)
	bytesSRTItalicEndTag            = []byte("</i>")
	bytesSRTItalicStartTag          = []byte("<i>")
	bytesSRTBoldEndTag              = []byte("</b>")
	bytesSRTBoldStartTag            = []byte("<b>")
	bytesSRTUnderlineEndTag         = []byte("</u>")
	bytesSRTUnderlineStartTag       = []byte("<u>")
)

// parseDurationSRT parses an .srt duration
func parseDurationSRT(i string) (time.Duration, error) {
	return parseDuration(i, ",", 3)
}

// ReadFromSRT parses an .srt content
func ReadFromSRT(i io.Reader) (o *Subtitles, err error) {
	// Init
	o = NewSubtitles()
	var scanner = bufio.NewScanner(i)

	// Scan
	var line string
	var lineNum int
	var s = &Item{}
	for scanner.Scan() {
		// Fetch line
		line = strings.TrimSpace(scanner.Text())
		lineNum++

		// Remove BOM header
		if lineNum == 1 {
			line = strings.TrimPrefix(line, string(BytesBOM))
		}

		// Line contains time boundaries
		if strings.Contains(line, srtTimeBoundariesSeparator) {
			// Return the wrong number of rows
			if len(s.Lines) == 0 {
				err = fmt.Errorf("astisub: line %d: no lines", lineNum)
				return
			}

			// Remove last item of previous subtitle since it's the index
			index := s.Lines[len(s.Lines)-1]
			s.Lines = s.Lines[:len(s.Lines)-1]

			// Remove trailing empty lines
			if len(s.Lines) > 0 {
				for i := len(s.Lines) - 1; i >= 0; i-- {
					if len(s.Lines[i].Items) > 0 {
						for j := len(s.Lines[i].Items) - 1; j >= 0; j-- {
							if len(s.Lines[i].Items[j].Text) == 0 {
								s.Lines[i].Items = s.Lines[i].Items[:j]
							} else {
								break
							}
						}
						if len(s.Lines[i].Items) == 0 {
							s.Lines = s.Lines[:i]
						}

					}
				}
			}

			// Init subtitle
			s = &Item{}

			// Fetch Index
			s.Index, _ = strconv.Atoi(index.String())

			// Extract time boundaries
			s1 := strings.Split(line, srtTimeBoundariesSeparator)
			if l := len(s1); l < 2 {
				err = fmt.Errorf("astisub: line %d: time boundaries has only %d element(s)", lineNum, l)
				return
			}
			// We do this to eliminate extra stuff like positions which are not documented anywhere
			s2 := strings.Split(s1[1], " ")

			// Parse time boundaries
			if s.StartAt, err = parseDurationSRT(s1[0]); err != nil {
				err = fmt.Errorf("astisub: line %d: parsing srt duration %s failed: %w", lineNum, s1[0], err)
				return
			}
			if s.EndAt, err = parseDurationSRT(s2[0]); err != nil {
				err = fmt.Errorf("astisub: line %d: parsing srt duration %s failed: %w", lineNum, s2[0], err)
				return
			}

			// Append subtitle
			o.Items = append(o.Items, s)
		} else {
			// Add text
			if l := parseTextSRT(line); len(l.Items) > 0 {
				s.Lines = append(s.Lines, l)
			}
		}
	}
	return
}

func parseTextSRT(i string) (o Line) {
	// Create tokenizer
	tr := html.NewTokenizer(strings.NewReader(i))

	// Loop
	italic := false
	bold := false
	underline := false
	for {
		// Get next tag
		t := tr.Next()

		// Process error
		if err := tr.Err(); err != nil {
			break
		}

		switch t {
		case html.EndTagToken:
			raw := tr.Raw()
			if bytes.Equal(raw, bytesSRTItalicEndTag) {
				italic = false
			} else if bytes.Equal(raw, bytesSRTBoldEndTag) {
				bold = false
			} else if bytes.Equal(raw, bytesSRTUnderlineEndTag) {
				underline = false
			}
		case html.StartTagToken:
			raw := tr.Raw()
			if bytes.Equal(raw, bytesSRTItalicStartTag) {
				italic = true
			} else if bytes.Equal(raw, bytesSRTBoldStartTag) {
				bold = true
			} else if bytes.Equal(raw, bytesSRTUnderlineStartTag) {
				underline = true
			}
		case html.TextToken:
			if s := strings.TrimSpace(string(tr.Raw())); s != "" {
				// Get style attribute
				var sa *StyleAttributes
				if italic || bold || underline {
					sa = &StyleAttributes{
						SRTItalic:    italic,
						SRTBold:      bold,
						SRTUnderline: underline,
					}
					sa.propagateSRTAttributes()
				}
				// Append item
				o.Items = append(o.Items, LineItem{
					InlineStyle: sa,
					Text:        s,
				})
			}
		}
	}

	return
}

// formatDurationSRT formats an .srt duration
func formatDurationSRT(i time.Duration) string {
	return formatDuration(i, ",", 3)
}

// WriteToSRT writes subtitles in .srt format
func (s Subtitles) WriteToSRT(o io.Writer) (err error) {
	// Do not write anything if no subtitles
	if len(s.Items) == 0 {
		err = ErrNoSubtitlesToWrite
		return
	}

	// Add BOM header
	var c []byte
	c = append(c, BytesBOM...)

	// Loop through subtitles
	for k, v := range s.Items {
		// Add time boundaries
		c = append(c, []byte(strconv.Itoa(k+1))...)
		c = append(c, bytesLineSeparator...)
		c = append(c, []byte(formatDurationSRT(v.StartAt))...)
		c = append(c, bytesSRTTimeBoundariesSeparator...)
		c = append(c, []byte(formatDurationSRT(v.EndAt))...)
		c = append(c, bytesLineSeparator...)

		// Loop through lines
		for _, l := range v.Lines {
			c = append(c, l.srtBytes()...)
		}

		// Add new line
		c = append(c, bytesLineSeparator...)
	}

	// Remove last new line
	c = c[:len(c)-1]

	// Write
	if _, err = o.Write(c); err != nil {
		err = fmt.Errorf("astisub: writing failed: %w", err)
		return
	}
	return
}

func (l Line) srtBytes() (c []byte) {
	for idx, li := range l.Items {
		c = append(c, li.srtBytes()...)
		// condition to avoid adding space as the last character.
		if idx < len(l.Items)-1 {
			c = append(c, []byte(" ")...)
		}
	}
	c = append(c, bytesLineSeparator...)
	return
}

func (li LineItem) srtBytes() (c []byte) {
	// Get styles
	i := li.InlineStyle != nil && li.InlineStyle.SRTItalic
	b := li.InlineStyle != nil && li.InlineStyle.SRTBold
	u := li.InlineStyle != nil && li.InlineStyle.SRTUnderline

	// Append
	if i {
		c = append(c, bytesSRTItalicStartTag...)
	}
	if b {
		c = append(c, bytesSRTBoldStartTag...)
	}
	if u {
		c = append(c, bytesSRTUnderlineStartTag...)
	}
	c = append(c, []byte(li.Text)...)
	if u {
		c = append(c, bytesSRTUnderlineEndTag...)
	}
	if b {
		c = append(c, bytesSRTBoldEndTag...)
	}
	if i {
		c = append(c, bytesSRTItalicEndTag...)
	}

	return
}
