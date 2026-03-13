package astisub

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/net/html"
)

// Constants
const (
	srtTimeBoundariesSeparator = "-->"
)

// Vars
var (
	bytesSRTTimeBoundariesSeparator = []byte(" " + srtTimeBoundariesSeparator + " ")
)

// parseDurationSRT parses an .srt duration
func parseDurationSRT(i string) (d time.Duration, err error) {
	for _, s := range []string{",", ".", ":"} {
		if d, err = parseDuration(i, s, 3); err == nil {
			return
		}
	}
	return
}

// ReadFromSRT parses an .srt content
func ReadFromSRT(i io.Reader) (o *Subtitles, err error) {
	// Init
	o = NewSubtitles()
	var scanner = newScanner(i)

	// Scan
	var line string
	var lineNum int
	var s = &Item{}
	var sa = &StyleAttributes{}
	for scanner.Scan() {
		// Fetch line
		line = strings.TrimSpace(scanner.Text())
		lineNum++
		if !utf8.ValidString(line) {
			err = fmt.Errorf("astisub: line %d is not valid utf-8", lineNum)
			return
		}

		// Remove BOM header
		if lineNum == 1 {
			line = strings.TrimPrefix(line, string(BytesBOM))
		}

		// Line contains time boundaries
		if strings.Contains(line, srtTimeBoundariesSeparator) {
			// Reset style attributes
			sa = &StyleAttributes{}

			// Remove last item of previous subtitle since it should be the index.
			// If the last line is empty then the item is missing an index.
			var index string
			if len(s.Lines) != 0 {
				index = s.Lines[len(s.Lines)-1].String()
				if index != "" {
					s.Lines = s.Lines[:len(s.Lines)-1]
				}
			}

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
			if index != "" {
				s.Index, _ = strconv.Atoi(index)
			}

			// Extract time boundaries
			s1 := strings.Split(line, srtTimeBoundariesSeparator)
			if l := len(s1); l < 2 {
				err = fmt.Errorf("astisub: line %d: time boundaries has only %d element(s)", lineNum, l)
				return
			}
			// We do this to eliminate extra stuff like positions which are not documented anywhere
			s2 := strings.Fields(s1[1])

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
			if l := parseTextSrt(line, sa); len(l.Items) > 0 {
				s.Lines = append(s.Lines, l)
			}
		}
	}
	return
}

// parseTextSrt parses the input line to fill the Line
func parseTextSrt(i string, sa *StyleAttributes) (o Line) {
	// special handling needed for empty line
	if strings.TrimSpace(i) == "" {
		o.Items = []LineItem{{Text: ""}}
		return
	}

	// Create tokenizer
	tr := html.NewTokenizer(strings.NewReader(i))

	// Loop
	for {
		// Get next tag
		t := tr.Next()

		// Process error
		if err := tr.Err(); err != nil {
			break
		}

		// Get unmodified text
		raw := string(tr.Raw())
		// Get current token
		token := tr.Token()

		switch t {
		case html.EndTagToken:
			// Parse italic/bold/underline
			switch token.Data {
			case "b":
				sa.SRTBold = false
			case "i":
				sa.SRTItalics = false
			case "u":
				sa.SRTUnderline = false
			case "font":
				sa.SRTColor = nil
			}
		case html.StartTagToken:
			// Parse italic/bold/underline
			switch token.Data {
			case "b":
				sa.SRTBold = true
			case "i":
				sa.SRTItalics = true
			case "u":
				sa.SRTUnderline = true
			case "font":
				if c := htmlTokenAttribute(&token, "color"); c != nil {
					// Parse the color string into a Color struct
					if color, err := newColorFromHTMLString(*c); err == nil {
						sa.SRTColor = color
					}
				}
			}
		case html.TextToken:
			if s := strings.TrimSpace(raw); s != "" {
				// Get style attribute
				var styleAttributes *StyleAttributes
				if sa.SRTBold || sa.SRTColor != nil || sa.SRTItalics || sa.SRTUnderline {
					styleAttributes = &StyleAttributes{
						SRTBold:      sa.SRTBold,
						SRTColor:     sa.SRTColor,
						SRTItalics:   sa.SRTItalics,
						SRTUnderline: sa.SRTUnderline,
					}
					styleAttributes.propagateSRTAttributes()
				}

				// Append item
				o.Items = append(o.Items, LineItem{
					InlineStyle: styleAttributes,
					Text:        unescapeHTML(raw),
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

	// Init writer
	w := bufio.NewWriter(o)
	defer w.Flush()

	// Add BOM header
	if _, err = w.Write(BytesBOM); err != nil {
		err = fmt.Errorf("astisub: writing bom failed: %w", err)
		return
	}

	// Loop through subtitles
	for k, v := range s.Items {
		// Add time boundaries
		if _, err = w.WriteString(strconv.Itoa(k + 1)); err != nil {
			err = fmt.Errorf("astisub: writing index failed: %w", err)
			return
		}
		if _, err = w.Write(bytesLineSeparator); err != nil {
			err = fmt.Errorf("astisub: writing line separator failed: %w", err)
			return
		}
		if _, err = w.WriteString(formatDurationSRT(v.StartAt)); err != nil {
			err = fmt.Errorf("astisub: writing start at failed: %w", err)
			return
		}
		if _, err = w.Write(bytesSRTTimeBoundariesSeparator); err != nil {
			err = fmt.Errorf("astisub: writing time boundaries separator failed: %w", err)
			return
		}
		if _, err = w.WriteString(formatDurationSRT(v.EndAt)); err != nil {
			err = fmt.Errorf("astisub: writing end at failed: %w", err)
			return
		}
		if _, err = w.Write(bytesLineSeparator); err != nil {
			err = fmt.Errorf("astisub: writing line separator failed: %w", err)
			return
		}

		// Loop through lines
		for _, l := range v.Lines {
			if err = l.writeSRT(w); err != nil {
				return
			}
		}

		// Add new line
		if k < len(s.Items)-1 {
			if _, err = w.Write(bytesLineSeparator); err != nil {
				err = fmt.Errorf("astisub: writing line separator failed: %w", err)
				return
			}
		}
	}
	return
}

func (l Line) writeSRT(w io.Writer) (err error) {
	for _, li := range l.Items {
		if err = li.writeSRT(w); err != nil {
			return
		}
	}
	if _, err = w.Write(bytesLineSeparator); err != nil {
		err = fmt.Errorf("astisub: writing line separator failed: %w", err)
		return
	}
	return
}

func (li LineItem) writeSRT(w io.Writer) (err error) {
	// Get color
	var color string
	if li.InlineStyle != nil && li.InlineStyle.SRTColor != nil {
		color = li.InlineStyle.SRTColor.HTMLString()
	}

	// Get bold/italics/underline
	b := li.InlineStyle != nil && li.InlineStyle.SRTBold
	i := li.InlineStyle != nil && li.InlineStyle.SRTItalics
	u := li.InlineStyle != nil && li.InlineStyle.SRTUnderline

	// Get position
	var pos byte
	if li.InlineStyle != nil {
		pos = li.InlineStyle.SRTPosition
	}

	// Append
	if color != "" {
		if _, err = w.Write([]byte("<font color=\"" + color + "\">")); err != nil {
			return fmt.Errorf("astisub: writing font color failed: %w", err)
		}
	}
	if b {
		if _, err = w.Write([]byte("<b>")); err != nil {
			return fmt.Errorf("astisub: writing bold failed: %w", err)
		}
	}
	if i {
		if _, err = w.Write([]byte("<i>")); err != nil {
			return fmt.Errorf("astisub: writing italics failed: %w", err)
		}
	}
	if u {
		if _, err = w.Write([]byte("<u>")); err != nil {
			return fmt.Errorf("astisub: writing underline failed: %w", err)
		}
	}
	if pos != 0 {
		if _, err = w.Write([]byte(fmt.Sprintf(`{\an%d}`, pos))); err != nil {
			return fmt.Errorf("astisub: writing position failed: %w", err)
		}
	}
	if _, err = w.Write([]byte(escapeHTML(li.Text))); err != nil {
		return fmt.Errorf("astisub: writing text failed: %w", err)
	}
	if u {
		if _, err = w.Write([]byte("</u>")); err != nil {
			return fmt.Errorf("astisub: writing underline close failed: %w", err)
		}
	}
	if i {
		if _, err = w.Write([]byte("</i>")); err != nil {
			return fmt.Errorf("astisub: writing italics close failed: %w", err)
		}
	}
	if b {
		if _, err = w.Write([]byte("</b>")); err != nil {
			return fmt.Errorf("astisub: writing bold close failed: %w", err)
		}
	}
	if color != "" {
		if _, err = w.Write([]byte("</font>")); err != nil {
			return fmt.Errorf("astisub: writing font close failed: %w", err)
		}
	}
	return
}
