package astisub

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"sort"

	"github.com/pkg/errors"
)

// https://www.w3.org/TR/webvtt1/

// Constants
const (
	webvttBlockNameComment        = "comment"
	webvttBlockNameRegion         = "region"
	webvttBlockNameStyle          = "style"
	webvttBlockNameText           = "text"
	webvttTimeBoundariesSeparator = " --> "
)

// Vars
var (
	bytesWebVTTTimeBoundariesSeparator = []byte(webvttTimeBoundariesSeparator)
)

// parseDurationWebVTT parses a .vtt duration
func parseDurationWebVTT(i string) (time.Duration, error) {
	return parseDuration(i, ".", 3)
}

// ReadFromWebVTT parses a .vtt content
// TODO Tags (u, i, b)
// TODO Class
// TODO Speaker name
func ReadFromWebVTT(i io.Reader) (o *Subtitles, err error) {
	// Init
	o = NewSubtitles()
	var scanner = bufio.NewScanner(i)
	var line string

	// Skip the header
	for scanner.Scan() {
		line = scanner.Text()
		line = strings.TrimPrefix(line, string(BytesBOM))
		if len(line) > 0 && strings.Fields(line)[0] == "WEBVTT" {
			break
		}
	}

	// Scan
	var item = &Item{}
	var blockName string
	var comments []string
	for scanner.Scan() {
		// Fetch line
		line = scanner.Text()
		// Check prefixes
		switch {
		// Comment
		case strings.HasPrefix(line, "NOTE "):
			blockName = webvttBlockNameComment
			comments = append(comments, strings.TrimPrefix(line, "NOTE "))
		// Empty line
		case len(line) == 0:
			// Reset block name
			blockName = ""
		// Region
		case strings.HasPrefix(line, "Region: "):
			// Add region styles
			var r = &Region{InlineStyle: &StyleAttributes{}}
			for _, part := range strings.Split(strings.TrimPrefix(line, "Region: "), " ") {
				// Split on "="
				var split = strings.Split(part, "=")
				if len(split) <= 1 {
					err = fmt.Errorf("astisub: Invalid region style %s", part)
					return
				}

				// Switch on key
				switch split[0] {
				case "id":
					r.ID = split[1]
				case "lines":
					if r.InlineStyle.WebVTTLines, err = strconv.Atoi(split[1]); err != nil {
						err = errors.Wrapf(err, "atoi of %s failed", split[1])
						return
					}
				case "regionanchor":
					r.InlineStyle.WebVTTRegionAnchor = split[1]
				case "scroll":
					r.InlineStyle.WebVTTScroll = split[1]
				case "viewportanchor":
					r.InlineStyle.WebVTTViewportAnchor = split[1]
				case "width":
					r.InlineStyle.WebVTTWidth = split[1]
				}
			}
			r.InlineStyle.propagateWebVTTAttributes()

			// Add region
			o.Regions[r.ID] = r
		// Style
		case strings.HasPrefix(line, "STYLE "):
			blockName = webvttBlockNameStyle
		// Time boundaries
		case strings.Contains(line, webvttTimeBoundariesSeparator):
			// Set block name
			blockName = webvttBlockNameText

			// Init new item
			item = &Item{
				Comments:    comments,
				InlineStyle: &StyleAttributes{},
			}

			// Split line on time boundaries
			var parts = strings.Split(line, webvttTimeBoundariesSeparator)
			// Split line on space to catch inline styles as well
			var partsRight = strings.Split(parts[1], " ")

			// Parse time boundaries
			if item.StartAt, err = parseDurationWebVTT(parts[0]); err != nil {
				err = errors.Wrapf(err, "astisub: parsing webvtt duration %s failed", parts[0])
				return
			}
			if item.EndAt, err = parseDurationWebVTT(partsRight[0]); err != nil {
				err = errors.Wrapf(err, "astisub: parsing webvtt duration %s failed", partsRight[0])
				return
			}

			// Parse style
			if len(partsRight) > 1 {
				// Add styles
				for index := 1; index < len(partsRight); index++ {
					// Split line on ":"
					var split = strings.Split(partsRight[index], ":")
					if len(split) <= 1 {
						err = fmt.Errorf("astisub: Invalid inline style %s", partsRight[index])
						return
					}

					// Switch on key
					switch split[0] {
					case "align":
						item.InlineStyle.WebVTTAlign = split[1]
					case "line":
						item.InlineStyle.WebVTTLine = split[1]
					case "position":
						item.InlineStyle.WebVTTPosition = split[1]
					case "region":
						if _, ok := o.Regions[split[1]]; !ok {
							err = fmt.Errorf("astisub: Unknown region %s", split[1])
							return
						}
						item.Region = o.Regions[split[1]]
					case "size":
						item.InlineStyle.WebVTTSize = split[1]
					case "vertical":
						item.InlineStyle.WebVTTVertical = split[1]
					}
				}
			}
			item.InlineStyle.propagateWebVTTAttributes()

			// Reset comments
			comments = []string{}

			// Append item
			o.Items = append(o.Items, item)
		// Text
		default:
			// Switch on block name
			switch blockName {
			case webvttBlockNameComment:
				comments = append(comments, line)
			case webvttBlockNameStyle:
				// TODO Do something with the style
			case webvttBlockNameText:
				item.Lines = append(item.Lines, Line{Items: []LineItem{{Text: line}}})
			default:
				// This is the ID
				// TODO Do something with the id
			}
		}
	}
	return
}

// formatDurationWebVTT formats a .vtt duration
func formatDurationWebVTT(i time.Duration) string {
	return formatDuration(i, ".", 3)
}

// WriteToWebVTT writes subtitles in .vtt format
func (s Subtitles) WriteToWebVTT(o io.Writer) (err error) {
	// Do not write anything if no subtitles
	if len(s.Items) == 0 {
		err = ErrNoSubtitlesToWrite
		return
	}

	// Add header
	var c []byte
	c = append(c, []byte("WEBVTT\n\n")...)

	// Add regions
	var k []string
	for _, region := range s.Regions {
		k = append(k, region.ID)
	}
	sort.Strings(k)
	for _, id := range k {
		c = append(c, []byte("Region: id="+s.Regions[id].ID)...)
		if s.Regions[id].InlineStyle.WebVTTLines != 0 {
			c = append(c, bytesSpace...)
			c = append(c, []byte("lines="+strconv.Itoa(s.Regions[id].InlineStyle.WebVTTLines))...)
		}
		if s.Regions[id].InlineStyle.WebVTTRegionAnchor != "" {
			c = append(c, bytesSpace...)
			c = append(c, []byte("regionanchor="+s.Regions[id].InlineStyle.WebVTTRegionAnchor)...)
		}
		if s.Regions[id].InlineStyle.WebVTTScroll != "" {
			c = append(c, bytesSpace...)
			c = append(c, []byte("scroll="+s.Regions[id].InlineStyle.WebVTTScroll)...)
		}
		if s.Regions[id].InlineStyle.WebVTTViewportAnchor != "" {
			c = append(c, bytesSpace...)
			c = append(c, []byte("viewportanchor="+s.Regions[id].InlineStyle.WebVTTViewportAnchor)...)
		}
		if s.Regions[id].InlineStyle.WebVTTWidth != "" {
			c = append(c, bytesSpace...)
			c = append(c, []byte("width="+s.Regions[id].InlineStyle.WebVTTWidth)...)
		}
		c = append(c, bytesLineSeparator...)
	}
	if len(s.Regions) > 0 {
		c = append(c, bytesLineSeparator...)
	}

	// Loop through subtitles
	for index, item := range s.Items {
		// Add comments
		if len(item.Comments) > 0 {
			c = append(c, []byte("NOTE ")...)
			for _, comment := range item.Comments {
				c = append(c, []byte(comment)...)
				c = append(c, bytesLineSeparator...)
			}
			c = append(c, bytesLineSeparator...)
		}

		// Add time boundaries
		c = append(c, []byte(strconv.Itoa(index+1))...)
		c = append(c, bytesLineSeparator...)
		c = append(c, []byte(formatDurationWebVTT(item.StartAt))...)
		c = append(c, bytesWebVTTTimeBoundariesSeparator...)
		c = append(c, []byte(formatDurationWebVTT(item.EndAt))...)

		// Add styles
		if item.InlineStyle != nil {
			if item.InlineStyle.WebVTTAlign != "" {
				c = append(c, bytesSpace...)
				c = append(c, []byte("align:"+item.InlineStyle.WebVTTAlign)...)
			}
			if item.InlineStyle.WebVTTLine != "" {
				c = append(c, bytesSpace...)
				c = append(c, []byte("line:"+item.InlineStyle.WebVTTLine)...)
			}
			if item.InlineStyle.WebVTTPosition != "" {
				c = append(c, bytesSpace...)
				c = append(c, []byte("position:"+item.InlineStyle.WebVTTPosition)...)
			}
			if item.Region != nil {
				c = append(c, bytesSpace...)
				c = append(c, []byte("region:"+item.Region.ID)...)
			}
			if item.InlineStyle.WebVTTSize != "" {
				c = append(c, bytesSpace...)
				c = append(c, []byte("size:"+item.InlineStyle.WebVTTSize)...)
			}
			if item.InlineStyle.WebVTTVertical != "" {
				c = append(c, bytesSpace...)
				c = append(c, []byte("vertical:"+item.InlineStyle.WebVTTVertical)...)
			}
		}

		// Add new line
		c = append(c, bytesLineSeparator...)

		// Loop through lines
		for _, l := range item.Lines {
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
