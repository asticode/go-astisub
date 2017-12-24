package astisub

import (
	"bufio"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Constants
const (
	srtTimeBoundariesSeparator = " --> "
)

// Vars
var (
	bytesSRTTimeBoundariesSeparator = []byte(srtTimeBoundariesSeparator)
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
	var s = &Item{}
	for scanner.Scan() {
		// Fetch line
		line = scanner.Text()

		// Line contains time boundaries
		if strings.Contains(line, srtTimeBoundariesSeparator) {
			// Remove last item of previous subtitle since it's the index
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

			// Fetch time boundaries
			boundaries := strings.Split(line, srtTimeBoundariesSeparator)
			if s.StartAt, err = parseDurationSRT(boundaries[0]); err != nil {
				err = errors.Wrapf(err, "astisub: parsing srt duration %s failed", boundaries[0])
				return
			}
			if s.EndAt, err = parseDurationSRT(boundaries[1]); err != nil {
				err = errors.Wrapf(err, "astisub: parsing srt duration %s failed", boundaries[1])
				return
			}

			// Append subtitle
			o.Items = append(o.Items, s)
		} else {
			// Add text
			s.Lines = append(s.Lines, Line{Items: []LineItem{{Text: line}}})
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
			c = append(c, []byte(l.String())...)
			c = append(c, bytesLineSeparator...)
		}

		// Add new line
		c = append(c, bytesLineSeparator...)
	}

	// Remove last new line
	c = c[:len(c)-1]

	// Write
	if _, err = o.Write(c); err != nil {
		err = errors.Wrap(err, "astisub: writing failed")
		return
	}
	return
}
