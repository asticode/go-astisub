package astisub

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Bytes
var (
	BytesBOM           = []byte{239, 187, 191}
	bytesLineSeparator = []byte("\n")
	bytesSpace         = []byte(" ")
)

// Errors
var (
	ErrInvalidExtension   = errors.New("Invalid extension")
	ErrNoSubtitlesToWrite = errors.New("No subtitles to write")
)

// Open opens a subtitle file
func Open(src string) (s *Subtitles, err error) {
	// Open the file
	var f *os.File
	if f, err = os.Open(src); err != nil {
		err = errors.Wrapf(err, "opening %s failed", src)
		return
	}
	defer f.Close()

	// Parse the content
	switch filepath.Ext(src) {
	case ".srt":
		s, err = ReadFromSRT(f)
	case ".stl":
		s, err = ReadFromSTL(f)
	case ".ttml":
		s, err = ReadFromTTML(f)
	case ".vtt":
		s, err = ReadFromWebVTT(f)
	default:
		err = ErrInvalidExtension
	}
	return
}

// Subtitles represents an ordered list of items with formatting
type Subtitles struct {
	Items    []*Item
	Metadata *Metadata
	Regions  []*Region
	Styles   []*Style
}

// Item represents a text to show between 2 time boundaries with formatting
type Item struct {
	Comments    []string
	EndAt       time.Duration
	InlineStyle *StyleAttributes
	Lines       []Line
	Region      *Region
	StartAt     time.Duration
	Style       *Style
}

// String implements the Stringer interface
func (i Item) String() string {
	var os []string
	for _, l := range i.Lines {
		os = append(os, l.String())
	}
	return strings.Join(os, " - ")
}

// StyleAttributes represents style attributes
// TODO Need more .ttml and .vtt style examples to get common patterns
type StyleAttributes struct {
	Align           string // WebVTT
	BackgroundColor string // TTML
	Color           string // TTML
	Direction       string // TTML
	Display         string // TTML
	DisplayAlign    string // TTML
	Extent          string // TTML
	FontFamily      string // TTML
	FontSize        string // TTML
	FontStyle       string // TTML
	FontWeight      string // TTML
	Line            string // WebVTT
	LineHeight      string // TTML
	Lines           int    // WebVTT
	Opacity         string // TTML
	Origin          string // TTML
	Overflow        string // TTML
	Padding         string // TTML
	Position        string // WebVTT
	RegionAnchor    string // WebVTT
	Scroll          string // WebVTT
	ShowBackground  string // TTML
	Size            string // WebVTT
	TextAlign       string // TTML
	TextDecoration  string // TTML
	TextOutline     string // TTML
	UnicodeBidi     string // TTML
	Vertical        string // WebVTT
	ViewportAnchor  string // WebVTT
	Visibility      string // TTML
	Width           string // WebVTT
	WrapOption      string // TTML
	WritingMode     string // TTML
	ZIndex          int    // TTML
}

// Metadata represents metadata
type Metadata struct {
	Copyright string
	Framerate int
	Language  string
	Title     string
}

// Region represents a subtitle's region
type Region struct {
	ID          string
	InlineStyle *StyleAttributes
	Style       *Style
}

// Style represents a subtitle's style
type Style struct {
	ID          string
	InlineStyle *StyleAttributes
	Style       *Style
}

// Line represents a set of formatted line items
type Line []LineItem

// String implement the Stringer interface
func (l Line) String() string {
	var texts []string
	for _, i := range l {
		texts = append(texts, i.Text)
	}
	return strings.Join(texts, " ")
}

// LineItem represents a formatted line item
type LineItem struct {
	InlineStyle *StyleAttributes
	Style       *Style
	Text        string
}

// Add adds a duration to each time boundaries. As in the time package, duration can be negative.
func (s *Subtitles) Add(d time.Duration) {
	for _, v := range s.Items {
		v.EndAt += d
		v.StartAt += d
	}
}

// Duration returns the subtitles duration
func (s Subtitles) Duration() time.Duration {
	if len(s.Items) == 0 {
		return time.Duration(0)
	}
	return s.Items[len(s.Items)-1].EndAt
}

// ForceDuration updates the subtitles duration.
// If requested duration is bigger, then we create a dummy item.
// If requested duration is smaller, then we remove useless items and we cut the last item or add a dummy item.
func (s *Subtitles) ForceDuration(d time.Duration) {
	// Requested duration is the same as the subtitles'one
	if s.Duration() == d {
		return
	}

	// Requested duration is bigger than subtitles'one
	if s.Duration() > d {
		// Find last item before input duration and update end at
		var lastIndex = -1
		for index, i := range s.Items {
			// Start at is bigger than input duration, we've found the last item
			if i.StartAt >= d {
				lastIndex = index
				break
			} else if i.EndAt > d {
				s.Items[index].EndAt = d
			}
		}

		// Last index has been found
		if lastIndex != -1 {
			s.Items = s.Items[:lastIndex]
		}
	}

	// Add dummy item
	if s.Duration() < d {
		s.Items = append(s.Items, &Item{EndAt: d, Lines: []Line{{{Text: "..."}}}, StartAt: d})
	}
}

// Fragment fragments subtitles with a specific fragment duration
func (s *Subtitles) Fragment(f time.Duration) {
	// Nothing to fragment
	if len(s.Items) == 0 {
		return
	}

	// Here we want to simulate fragments of duration f until there are no subtitles left in that period of time
	var fragmentStartAt, fragmentEndAt = time.Duration(0), f
	for fragmentStartAt < s.Items[len(s.Items)-1].EndAt {
		// We loop through subtitles and process the ones that either contain the fragment start at,
		// or contain the fragment end at
		//
		// It's useless processing subtitles contained between fragment start at and end at
		//             |____________________|             <- subtitle
		//           |                        |
		//   fragment start at        fragment end at
		for i, sub := range s.Items {
			// Init
			var newSub = &Item{}
			*newSub = *sub

			// A switch is more readable here
			switch {
			// Subtitle contains fragment start at
			// |____________________|                         <- subtitle
			//           |                        |
			//   fragment start at        fragment end at
			case sub.StartAt < fragmentStartAt && sub.EndAt > fragmentStartAt:
				sub.StartAt = fragmentStartAt
				newSub.EndAt = fragmentStartAt
			// Subtitle contains fragment end at
			//                         |____________________| <- subtitle
			//           |                        |
			//   fragment start at        fragment end at
			case sub.StartAt < fragmentEndAt && sub.EndAt > fragmentEndAt:
				sub.StartAt = fragmentEndAt
				newSub.EndAt = fragmentEndAt
			default:
				continue
			}

			// Insert new sub
			s.Items = append(s.Items[:i], append([]*Item{newSub}, s.Items[i:]...)...)
		}

		// Update fragments boundaries
		fragmentStartAt += f
		fragmentEndAt += f
	}
}

// IsEmpty returns whether the subtitles are empty
func (s Subtitles) IsEmpty() bool {
	return len(s.Items) == 0
}

// Merge merges subtitles i into subtitles
func (s *Subtitles) Merge(i *Subtitles) {
	// Append
	s.Items = append(s.Items, i.Items...)
	s.Regions = append(s.Regions, i.Regions...)
	s.Styles = append(s.Styles, i.Styles...)

	// Order items
	s.Order()
}

// Order orders items
func (s *Subtitles) Order() {
	// Nothing to do if less than 1 element
	if len(s.Items) <= 1 {
		return
	}

	// Order
	var swapped = true
	for swapped {
		swapped = false
		for index := 1; index < len(s.Items); index++ {
			if s.Items[index-1].StartAt > s.Items[index].StartAt {
				var tmp = s.Items[index-1]
				s.Items[index-1] = s.Items[index]
				s.Items[index] = tmp
				swapped = true
			}
		}
	}
}

// Unfragment unfragments subtitles
func (s *Subtitles) Unfragment() {
	// Nothing to do if less than 1 element
	if len(s.Items) <= 1 {
		return
	}

	// Loop through items
	var previousItem = s.Items[0]
	for index := 1; index < len(s.Items); index++ {
		// Items are the same
		if s.Items[index].String() == previousItem.String() && previousItem.EndAt == s.Items[index].StartAt {
			previousItem.EndAt = s.Items[index].EndAt
			s.Items = append(s.Items[:index], s.Items[index+1:]...)
			index--
		} else {
			previousItem = s.Items[index]
		}
	}
}

// Write writes subtitles to a file
func (s Subtitles) Write(dst string) (err error) {
	// Create the file
	var f *os.File
	if f, err = os.Create(dst); err != nil {
		err = errors.Wrapf(err, "creating %s failed", dst)
		return
	}
	defer f.Close()

	// Write the content
	switch filepath.Ext(dst) {
	case ".srt":
		err = s.WriteToSRT(f)
	case ".stl":
		err = s.WriteToSTL(f)
	case ".ttml":
		err = s.WriteToTTML(f)
	case ".vtt":
		err = s.WriteToWebVTT(f)
	default:
		err = ErrInvalidExtension
	}
	return
}

// parseDuration parses a duration in "00:00:00.000" or "00:00:00,000" format
func parseDuration(i, millisecondSep string) (o time.Duration, err error) {
	// Split milliseconds
	var parts = strings.Split(i, millisecondSep)
	var milliseconds int
	var s string
	if len(parts) >= 2 {
		// Invalid number of millisecond digits
		if len(parts[1]) > 3 {
			err = fmt.Errorf("Invalid number of millisecond digits detected in %s", i)
			return
		}

		// Parse milliseconds
		s = strings.TrimSpace(parts[1])
		if milliseconds, err = strconv.Atoi(s); err != nil {
			err = errors.Wrapf(err, "atoi of %s failed", s)
			return
		}

		// In case number of milliseconds digits is not 3
		if len(s) == 2 {
			milliseconds *= 10
		} else if len(s) == 1 {
			milliseconds *= 100
		}
	}

	// Split hours, minutes and seconds
	parts = strings.Split(strings.TrimSpace(parts[0]), ":")
	var partSeconds, partMinutes, partHours string
	if len(parts) == 2 {
		partSeconds = parts[1]
		partMinutes = parts[0]
	} else if len(parts) == 3 {
		partSeconds = parts[2]
		partMinutes = parts[1]
		partHours = parts[0]
	} else {
		err = fmt.Errorf("No hours, minutes or seconds detected in %s", i)
		return
	}

	// Parse seconds
	var seconds int
	s = strings.TrimSpace(partSeconds)
	if seconds, err = strconv.Atoi(s); err != nil {
		err = errors.Wrapf(err, "atoi of %s failed", s)
		return
	}

	// Parse minutes
	var minutes int
	s = strings.TrimSpace(partMinutes)
	if minutes, err = strconv.Atoi(s); err != nil {
		err = errors.Wrapf(err, "atoi of %s failed", s)
		return
	}

	// Parse hours
	var hours int
	if len(partHours) > 0 {
		s = strings.TrimSpace(partHours)
		if hours, err = strconv.Atoi(s); err != nil {
			err = errors.Wrapf(err, "atoi of %s failed", s)
			return
		}
	}

	// Generate output
	o = time.Duration(milliseconds)*time.Millisecond + time.Duration(seconds)*time.Second + time.Duration(minutes)*time.Minute + time.Duration(hours)*time.Hour
	return
}

// formatDurationSRT formats a duration
func formatDuration(i time.Duration, millisecondSep string) (s string) {
	// Parse hours
	var hours = int(i / time.Hour)
	var n = i % time.Hour
	if hours < 10 {
		s += "0"
	}
	s += strconv.Itoa(hours) + ":"

	// Parse minutes
	var minutes = int(n / time.Minute)
	n = i % time.Minute
	if minutes < 10 {
		s += "0"
	}
	s += strconv.Itoa(minutes) + ":"

	// Parse seconds
	var seconds = int(n / time.Second)
	n = i % time.Second
	if seconds < 10 {
		s += "0"
	}
	s += strconv.Itoa(seconds) + millisecondSep

	// Parse milliseconds
	var milliseconds = int(n / time.Millisecond)
	if milliseconds < 10 {
		s += "00"
	} else if milliseconds < 100 {
		s += "0"
	}
	s += strconv.Itoa(milliseconds)
	return
}
