package astisub

import (
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/net/html"
)

// https://www.w3.org/TR/webvtt1/

// Constants
const (
	webvttBlockNameComment        = "comment"
	webvttBlockNameRegion         = "region"
	webvttBlockNameStyle          = "style"
	webvttBlockNameText           = "text"
	webvttDefaultStyleID          = "astisub-webvtt-default-style-id"
	webvttTimeBoundariesSeparator = "-->"
	webvttTimestampMapHeader      = "X-TIMESTAMP-MAP"
)

// Vars
var (
	bytesWebVTTItalicEndTag            = []byte("</i>")
	bytesWebVTTItalicStartTag          = []byte("<i>")
	bytesWebVTTTimeBoundariesSeparator = []byte(" " + webvttTimeBoundariesSeparator + " ")
	webVTTRegexpInlineTimestamp        = regexp.MustCompile(`<((?:\d{2,}:)?\d{2}:\d{2}\.\d{3})>`)
	webVTTRegexpTag                    = regexp.MustCompile(`(</*\s*([^\.\s]+)(\.[^\s/]*)*\s*([^/]*)\s*/*>)`)
)

// parseDurationWebVTT parses a .vtt duration
func parseDurationWebVTT(i string) (time.Duration, error) {
	return parseDuration(i, ".", 3)
}

// WebVTTTimestampMap is a structure for storing timestamps for WEBVTT's
// X-TIMESTAMP-MAP feature commonly used for syncing cue times with
// MPEG-TS streams.
type WebVTTTimestampMap struct {
	Local  time.Duration
	MpegTS int64
}

// Offset calculates and returns the time offset described by the
// timestamp map.
func (t *WebVTTTimestampMap) Offset() time.Duration {
	if t == nil {
		return 0
	}
	return time.Duration(t.MpegTS)*time.Second/90000 - t.Local
}

// String implements Stringer interface for TimestampMap, returning
// the fully formatted header string for the instance.
func (t *WebVTTTimestampMap) String() string {
	mpegts := fmt.Sprintf("MPEGTS:%d", t.MpegTS)
	local := fmt.Sprintf("LOCAL:%s", formatDurationWebVTT(t.Local))
	return fmt.Sprintf("%s=%s,%s", webvttTimestampMapHeader, local, mpegts)
}

// https://tools.ietf.org/html/rfc8216#section-3.5
// Eg., `X-TIMESTAMP-MAP=LOCAL:00:00:00.000,MPEGTS:900000` => 10s
//
//	`X-TIMESTAMP-MAP=LOCAL:00:00:00.000,MPEGTS:180000` => 2s
func parseWebVTTTimestampMap(line string) (timestampMap *WebVTTTimestampMap, err error) {
	splits := strings.Split(line, "=")
	if len(splits) <= 1 {
		err = fmt.Errorf("astisub: invalid X-TIMESTAMP-MAP, no '=' found")
		return
	}
	right := splits[1]

	var local time.Duration
	var mpegts int64
	for _, split := range strings.Split(right, ",") {
		splits := strings.SplitN(split, ":", 2)
		if len(splits) <= 1 {
			err = fmt.Errorf("astisub: invalid X-TIMESTAMP-MAP, part %q didn't contain ':'", right)
			return
		}

		switch strings.ToLower(strings.TrimSpace(splits[0])) {
		case "local":
			local, err = parseDurationWebVTT(splits[1])
			if err != nil {
				err = fmt.Errorf("astisub: parsing webvtt duration failed: %w", err)
				return
			}
		case "mpegts":
			mpegts, err = strconv.ParseInt(splits[1], 10, 0)
			if err != nil {
				err = fmt.Errorf("astisub: parsing int %s failed: %w", splits[1], err)
				return
			}
		}
	}

	timestampMap = &WebVTTTimestampMap{
		Local:  local,
		MpegTS: mpegts,
	}
	return
}

// ReadFromWebVTT parses a .vtt content
// TODO Tags (u, i, b)
// TODO Class
func ReadFromWebVTT(i io.Reader) (o *Subtitles, err error) {
	// Init
	o = NewSubtitles()
	var scanner = newScanner(i)

	var line string
	var lineNum int

	// Skip the header
	for scanner.Scan() {
		lineNum++
		line = scanner.Text()
		line = strings.TrimPrefix(line, string(BytesBOM))
		if !utf8.ValidString(line) {
			err = fmt.Errorf("astisub: line %d is not valid utf-8", lineNum)
			return
		}
		if fs := strings.Fields(line); len(fs) > 0 && fs[0] == "WEBVTT" {
			break
		}
	}

	// Scan
	var item = &Item{}
	var blockName string
	var comments []string
	var index int
	var sa = &StyleAttributes{}

	for scanner.Scan() {
		// Fetch line
		line = strings.TrimSpace(scanner.Text())
		lineNum++
		if !utf8.ValidString(line) {
			err = fmt.Errorf("astisub: line %d is not valid utf-8", lineNum)
			return
		}

		switch {
		// Comment
		case strings.HasPrefix(line, "NOTE "):
			blockName = webvttBlockNameComment
			comments = append(comments, strings.TrimPrefix(line, "NOTE "))
		// Empty line
		case len(line) == 0:
			// Reset block name, if we are not in the middle of CSS.
			// If we are in STYLE block and the CSS is empty or we meet the right brace at the end of last line,
			// then we are not in CSS and can switch to parse next WebVTT block.
			if blockName != webvttBlockNameStyle || sa == nil ||
				len(sa.WebVTTStyles) == 0 ||
				strings.HasSuffix(sa.WebVTTStyles[len(sa.WebVTTStyles)-1], "}") {
				blockName = ""
			}

			// Reset WebVTTTags
			sa.WebVTTTags = []WebVTTTag{}

		// Region
		case strings.HasPrefix(line, "Region: "):
			// Add region styles
			var r = &Region{InlineStyle: &StyleAttributes{}}
			for _, part := range strings.Split(strings.TrimPrefix(line, "Region: "), " ") {
				// Split on "="
				var split = strings.Split(part, "=")
				if len(split) <= 1 {
					err = fmt.Errorf("astisub: line %d: Invalid region style %s", lineNum, part)
					return
				}

				// Switch on key
				switch split[0] {
				case "id":
					r.ID = split[1]
				case "lines":
					if r.InlineStyle.WebVTTLines, err = strconv.Atoi(split[1]); err != nil {
						err = fmt.Errorf("atoi of %s failed: %w", split[1], err)
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
		case strings.HasPrefix(line, "STYLE"):
			blockName = webvttBlockNameStyle

			if _, ok := o.Styles[webvttDefaultStyleID]; !ok {
				sa = &StyleAttributes{}
				o.Styles[webvttDefaultStyleID] = &Style{
					InlineStyle: sa,
					ID:          webvttDefaultStyleID,
				}
			}

		// Time boundaries
		case strings.Contains(line, webvttTimeBoundariesSeparator):
			// Set block name
			blockName = webvttBlockNameText

			// Init new item
			item = &Item{
				Comments:    comments,
				Index:       index,
				InlineStyle: &StyleAttributes{},
			}

			// Reset index
			index = 0

			// Split line on time boundaries
			var left = strings.Split(line, webvttTimeBoundariesSeparator)

			// Split line on space to get remaining of time data
			var right = strings.Fields(left[1])

			// Parse time boundaries
			if item.StartAt, err = parseDurationWebVTT(left[0]); err != nil {
				err = fmt.Errorf("astisub: line %d: parsing webvtt duration %s failed: %w", lineNum, left[0], err)
				return
			}
			if item.EndAt, err = parseDurationWebVTT(right[0]); err != nil {
				err = fmt.Errorf("astisub: line %d: parsing webvtt duration %s failed: %w", lineNum, right[0], err)
				return
			}

			// Parse style
			if len(right) > 1 {
				// Add styles
				for index := 1; index < len(right); index++ {
					// Empty
					if right[index] == "" {
						continue
					}

					// Split line on ":"
					var split = strings.Split(right[index], ":")
					if len(split) <= 1 {
						err = fmt.Errorf("astisub: line %d: Invalid inline style '%s'", lineNum, right[index])
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
							err = fmt.Errorf("astisub: line %d: Unknown region %s", lineNum, split[1])
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

		case strings.HasPrefix(line, webvttTimestampMapHeader):
			if len(item.Lines) > 0 {
				err = errors.New("astisub: found timestamp map after processing subtitle items")
				return
			}

			var timestampMap *WebVTTTimestampMap
			timestampMap, err = parseWebVTTTimestampMap(line)
			if err != nil {
				err = fmt.Errorf("astisub: parsing webvtt timestamp map failed: %w", err)
				return
			}
			if o.Metadata == nil {
				o.Metadata = new(Metadata)
			}
			o.Metadata.WebVTTTimestampMap = timestampMap

		// Text
		default:
			// Switch on block name
			switch blockName {
			case webvttBlockNameComment:
				comments = append(comments, line)
			case webvttBlockNameStyle:
				sa.WebVTTStyles = append(sa.WebVTTStyles, line)
			case webvttBlockNameText:
				// Parse line
				if l := parseTextWebVTT(line, sa); len(l.Items) > 0 {
					item.Lines = append(item.Lines, l)
				}
			default:
				// This is the ID
				index, _ = strconv.Atoi(line)
			}
		}
	}
	return
}

// parseTextWebVTT parses the input line to fill the Line
func parseTextWebVTT(i string, sa *StyleAttributes) (o Line) {
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

		switch t {
		case html.EndTagToken:
			// Pop the top of stack if we meet end tag
			if len(sa.WebVTTTags) > 0 {
				sa.WebVTTTags = sa.WebVTTTags[:len(sa.WebVTTTags)-1]
			}
		case html.StartTagToken:
			if matches := webVTTRegexpTag.FindStringSubmatch(string(tr.Raw())); len(matches) > 4 {
				tagName := matches[2]

				var classes []string
				if matches[3] != "" {
					classes = strings.Split(strings.Trim(matches[3], "."), ".")
				}

				annotation := ""
				if matches[4] != "" {
					annotation = strings.TrimSpace(matches[4])
				}

				if tagName == "v" {
					if o.VoiceName == "" {
						// Only get voicename of the first <v> appears in the line
						o.VoiceName = annotation
					} else {
						// TODO: do something with other <v> instead of ignoring
						log.Printf("astisub: found another voice name %q in %q. Ignore", annotation, i)
					}
					continue
				}

				// Push the tag to stack
				sa.WebVTTTags = append(sa.WebVTTTags, WebVTTTag{
					Name:       tagName,
					Classes:    classes,
					Annotation: annotation,
				})
			}

		case html.TextToken:
			// Get style attribute
			var styleAttributes *StyleAttributes
			if len(sa.WebVTTTags) > 0 {
				tags := make([]WebVTTTag, len(sa.WebVTTTags))
				copy(tags, sa.WebVTTTags)
				styleAttributes = &StyleAttributes{
					WebVTTTags: tags,
				}
				styleAttributes.propagateWebVTTAttributes()
			}

			// Append items
			o.Items = append(o.Items, parseTextWebVTTTextToken(styleAttributes, string(tr.Raw()))...)
		}
	}
	return
}

func parseTextWebVTTTextToken(sa *StyleAttributes, line string) (ret []LineItem) {
	// split the line by inline timestamps
	indexes := webVTTRegexpInlineTimestamp.FindAllStringSubmatchIndex(line, -1)

	if len(indexes) == 0 {
		return []LineItem{{
			InlineStyle: sa,
			Text:        unescapeHTML(line),
		}}
	}

	// get the text before the first timestamp
	if s := line[:indexes[0][0]]; strings.TrimSpace(s) != "" {
		ret = append(ret, LineItem{
			InlineStyle: sa,
			Text:        unescapeHTML(s),
		})
	}

	for i, match := range indexes {
		// get the text between the timestamps
		endIndex := len(line)
		if i+1 < len(indexes) {
			endIndex = indexes[i+1][0]
		}
		s := line[match[1]:endIndex]
		if strings.TrimSpace(s) == "" {
			continue
		}

		// Parse timestamp
		t, err := parseDurationWebVTT(line[match[2]:match[3]])
		if err != nil {
			log.Printf("astisub: parsing webvtt duration %s failed, ignoring: %v", line[match[2]:match[3]], err)
		}

		ret = append(ret, LineItem{
			InlineStyle: sa,
			StartAt:     t,
			Text:        unescapeHTML(s),
		})
	}

	return
}

// formatDurationWebVTT formats a .vtt duration
func formatDurationWebVTT(i time.Duration) string {
	return formatDuration(i, ".", 3)
}

// WriteToWebVTT writes subtitles in .vtt format
// if set true in second args write index as item index
func (s Subtitles) WriteToWebVTT(args ...interface{}) (err error) {
	var o io.Writer
	writeWithIndex := false
	for i, arg := range args {
		switch i {
		case 0: // default output writer
			out, ok := arg.(io.Writer)
			if !ok {
				return fmt.Errorf("first input argument must be io.Writer")
			}
			o = out
		case 1:
			b, ok := arg.(bool)
			if !ok {
				return fmt.Errorf("second input argument must be boolean")
			}
			writeWithIndex = b
		}
	}
	// Do not write anything if no subtitles
	if len(s.Items) == 0 {
		err = ErrNoSubtitlesToWrite
		return
	}

	// Add header
	var c []byte
	c = append(c, []byte("WEBVTT")...)

	// Write X-TIMESTAMP-MAP if set
	if s.Metadata != nil {
		webVTTTimestampMap := s.Metadata.WebVTTTimestampMap
		if webVTTTimestampMap != nil {
			c = append(c, []byte("\n")...)
			c = append(c, []byte(webVTTTimestampMap.String())...)
		}
	}
	c = append(c, []byte("\n\n")...)

	var style []string
	for _, s := range s.Styles {
		if s.InlineStyle != nil {
			style = append(style, s.InlineStyle.WebVTTStyles...)
		}
	}

	if len(style) > 0 {
		c = append(c, []byte(fmt.Sprintf("STYLE\n%s\n\n", strings.Join(style, "\n")))...)
	}

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
		} else if s.Regions[id].Style != nil && s.Regions[id].Style.InlineStyle != nil && s.Regions[id].Style.InlineStyle.WebVTTLines != 0 {
			c = append(c, bytesSpace...)
			c = append(c, []byte("lines="+strconv.Itoa(s.Regions[id].Style.InlineStyle.WebVTTLines))...)
		}
		if s.Regions[id].InlineStyle.WebVTTRegionAnchor != "" {
			c = append(c, bytesSpace...)
			c = append(c, []byte("regionanchor="+s.Regions[id].InlineStyle.WebVTTRegionAnchor)...)
		} else if s.Regions[id].Style != nil && s.Regions[id].Style.InlineStyle != nil && s.Regions[id].Style.InlineStyle.WebVTTRegionAnchor != "" {
			c = append(c, bytesSpace...)
			c = append(c, []byte("regionanchor="+s.Regions[id].Style.InlineStyle.WebVTTRegionAnchor)...)
		}
		if s.Regions[id].InlineStyle.WebVTTScroll != "" {
			c = append(c, bytesSpace...)
			c = append(c, []byte("scroll="+s.Regions[id].InlineStyle.WebVTTScroll)...)
		} else if s.Regions[id].Style != nil && s.Regions[id].Style.InlineStyle != nil && s.Regions[id].Style.InlineStyle.WebVTTScroll != "" {
			c = append(c, bytesSpace...)
			c = append(c, []byte("scroll="+s.Regions[id].Style.InlineStyle.WebVTTScroll)...)
		}
		if s.Regions[id].InlineStyle.WebVTTViewportAnchor != "" {
			c = append(c, bytesSpace...)
			c = append(c, []byte("viewportanchor="+s.Regions[id].InlineStyle.WebVTTViewportAnchor)...)
		} else if s.Regions[id].Style != nil && s.Regions[id].Style.InlineStyle != nil && s.Regions[id].Style.InlineStyle.WebVTTViewportAnchor != "" {
			c = append(c, bytesSpace...)
			c = append(c, []byte("viewportanchor="+s.Regions[id].Style.InlineStyle.WebVTTViewportAnchor)...)
		}
		if s.Regions[id].InlineStyle.WebVTTWidth != "" {
			c = append(c, bytesSpace...)
			c = append(c, []byte("width="+s.Regions[id].InlineStyle.WebVTTWidth)...)
		} else if s.Regions[id].Style != nil && s.Regions[id].Style.InlineStyle != nil && s.Regions[id].Style.InlineStyle.WebVTTWidth != "" {
			c = append(c, bytesSpace...)
			c = append(c, []byte("width="+s.Regions[id].Style.InlineStyle.WebVTTWidth)...)
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
		if writeWithIndex {
			c = append(c, []byte(strconv.Itoa(item.Index))...)
		} else {
			c = append(c, []byte(strconv.Itoa(index+1))...)
		}
		c = append(c, bytesLineSeparator...)
		c = append(c, []byte(formatDurationWebVTT(item.StartAt))...)
		c = append(c, bytesWebVTTTimeBoundariesSeparator...)
		c = append(c, []byte(formatDurationWebVTT(item.EndAt))...)

		// Add styles
		if item.InlineStyle != nil {
			if item.InlineStyle.WebVTTAlign != "" {
				c = append(c, bytesSpace...)
				c = append(c, []byte("align:"+item.InlineStyle.WebVTTAlign)...)
			} else if item.Style != nil && item.Style.InlineStyle != nil && item.Style.InlineStyle.WebVTTAlign != "" {
				c = append(c, bytesSpace...)
				c = append(c, []byte("align:"+item.Style.InlineStyle.WebVTTAlign)...)
			}
			if item.InlineStyle.WebVTTLine != "" {
				c = append(c, bytesSpace...)
				c = append(c, []byte("line:"+item.InlineStyle.WebVTTLine)...)
			} else if item.Style != nil && item.Style.InlineStyle != nil && item.Style.InlineStyle.WebVTTLine != "" {
				c = append(c, bytesSpace...)
				c = append(c, []byte("line:"+item.Style.InlineStyle.WebVTTLine)...)
			}
			if item.InlineStyle.WebVTTPosition != "" {
				c = append(c, bytesSpace...)
				c = append(c, []byte("position:"+item.InlineStyle.WebVTTPosition)...)
			} else if item.Style != nil && item.Style.InlineStyle != nil && item.Style.InlineStyle.WebVTTPosition != "" {
				c = append(c, bytesSpace...)
				c = append(c, []byte("position:"+item.Style.InlineStyle.WebVTTPosition)...)
			}
			if item.Region != nil {
				c = append(c, bytesSpace...)
				c = append(c, []byte("region:"+item.Region.ID)...)
			}
			if item.InlineStyle.WebVTTSize != "" {
				c = append(c, bytesSpace...)
				c = append(c, []byte("size:"+item.InlineStyle.WebVTTSize)...)
			} else if item.Style != nil && item.Style.InlineStyle != nil && item.Style.InlineStyle.WebVTTSize != "" {
				c = append(c, bytesSpace...)
				c = append(c, []byte("size:"+item.Style.InlineStyle.WebVTTSize)...)
			}
			if item.InlineStyle.WebVTTVertical != "" {
				c = append(c, bytesSpace...)
				c = append(c, []byte("vertical:"+item.InlineStyle.WebVTTVertical)...)
			} else if item.Style != nil && item.Style.InlineStyle != nil && item.Style.InlineStyle.WebVTTVertical != "" {
				c = append(c, bytesSpace...)
				c = append(c, []byte("vertical:"+item.Style.InlineStyle.WebVTTVertical)...)
			}
		}

		// Add new line
		c = append(c, bytesLineSeparator...)

		// Loop through lines
		for _, l := range item.Lines {
			c = append(c, l.webVTTBytes()...)
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

func (l Line) webVTTBytes() (c []byte) {
	if l.VoiceName != "" {
		c = append(c, []byte("<v "+l.VoiceName+">")...)
	}
	for idx := 0; idx < len(l.Items); idx++ {
		var previous, next *LineItem
		if idx > 0 {
			previous = &l.Items[idx-1]
		}
		if idx < len(l.Items)-1 {
			next = &l.Items[idx+1]
		}
		c = append(c, l.Items[idx].webVTTBytes(previous, next)...)
	}
	c = append(c, bytesLineSeparator...)
	return
}

func (li LineItem) webVTTBytes(previous, next *LineItem) (c []byte) {
	// Add timestamp
	if li.StartAt > 0 {
		c = append(c, []byte("<"+formatDurationWebVTT(li.StartAt)+">")...)
	}

	// Get color
	var color string
	if li.InlineStyle != nil && li.InlineStyle.TTMLColor != nil {
		color = cssColor(*li.InlineStyle.TTMLColor)
	}

	// Append
	if color != "" {
		c = append(c, []byte("<c."+color+">")...)
	}
	if li.InlineStyle != nil {
		for idx, tag := range li.InlineStyle.WebVTTTags {
			if previous != nil && previous.InlineStyle != nil && len(previous.InlineStyle.WebVTTTags) > idx && tag.Name == previous.InlineStyle.WebVTTTags[idx].Name {
				continue
			}
			c = append(c, []byte(tag.startTag())...)
		}
	}
	c = append(c, []byte(escapeHTML(li.Text))...)
	if li.InlineStyle != nil {
		for i := len(li.InlineStyle.WebVTTTags) - 1; i >= 0; i-- {
			tag := li.InlineStyle.WebVTTTags[i]
			if next != nil && next.InlineStyle != nil && len(next.InlineStyle.WebVTTTags) > i && tag.Name == next.InlineStyle.WebVTTTags[i].Name {
				continue
			}
			c = append(c, []byte(tag.endTag())...)
		}
	}
	if color != "" {
		c = append(c, []byte("</c>")...)
	}
	return
}

func cssColor(rgb string) string {
	colors := map[string]string{
		"#00ffff": "cyan",    // narrator, thought
		"#ffff00": "yellow",  // out of vision
		"#ff0000": "red",     // noises
		"#ff00ff": "magenta", // song
		"#00ff00": "lime",    // foreign speak
	}
	return colors[strings.ToLower(rgb)] // returning the empty string is ok
}
