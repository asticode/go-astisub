package astisub

import (
	"regexp"
	"strconv"
	"strings"
)

// http://docs.aegisub.org/latest/ASS_Tags/
// https://github.com/libass/libass

// Override tags such as {\i1} or {\an8} are not part of the original SubStation Alpha
// specification: they come from its v4.00+ extension, which has no formal spec and is
// only documented by the implementations everyone follows: Aegisub and libass. A tag is
// made of a backslash, a name and its arguments, the arguments being either
// parenthesized or running until the next tag. Anything else inside the braces is a
// comment, which is why only the leading number or color of an argument is read:
// {\i1 this is ignored\b1}.

// SSA override tag regexps
var (
	ssaRegexpOverrideBlock  = regexp.MustCompile(`\{\s*\\[^}]*\}`)
	ssaRegexpOverrideColor  = regexp.MustCompile(`^(&H)?([0-9a-fA-F]+)`)
	ssaRegexpOverrideNumber = regexp.MustCompile(`^[-+]?[0-9]*\.?[0-9]*`)
	ssaRegexpOverrideTag    = regexp.MustCompile(`\\([0-9]?[a-zA-Z]+)(\([^)]*\)|[^\\]*)`)
)

// ssaOverrideTag represents an override tag, e.g. `\an8` gives a tag named "an" with the
// arguments "8". A digit prefixes the name of the tags selecting one of the four colors,
// e.g. `\3a`.
type ssaOverrideTag struct {
	name string
	args string
}

// ssaOverridePart represents a piece of a text, either an override block or literal text
type ssaOverridePart struct {
	tags []ssaOverrideTag // set when the part is an override block
	text string           // set otherwise
}

// splitSSAOverrides splits a text into its literal parts and its override blocks, the
// tags of which are parsed
func splitSSAOverrides(t string) (o []ssaOverridePart) {
	var previous int
	for _, idxs := range ssaRegexpOverrideBlock.FindAllStringIndex(t, -1) {
		if idxs[0] > previous {
			o = append(o, ssaOverridePart{text: t[previous:idxs[0]]})
		}

		var tags []ssaOverrideTag
		for _, m := range ssaRegexpOverrideTag.FindAllStringSubmatch(t[idxs[0]+1:idxs[1]-1], -1) {
			tags = append(tags, ssaOverrideTag{name: m[1], args: strings.TrimSpace(m[2])})
		}
		o = append(o, ssaOverridePart{tags: tags})
		previous = idxs[1]
	}
	if previous < len(t) {
		o = append(o, ssaOverridePart{text: t[previous:]})
	}
	return
}

// applySRTOverrideTags updates the srt style attributes with the tags. The tags srt
// cannot represent (positioning, karaoke, animations, fonts, ...) are ignored
func applySRTOverrideTags(sa *StyleAttributes, tags []ssaOverrideTag) {
	for _, tag := range tags {
		switch tag.name {
		case "r": // reverts every tag
			*sa = StyleAttributes{}
		case "a", "an":
			sa.SRTPosition = ssaOverrideAlignment(tag, sa.SRTPosition)
		case "b":
			sa.SRTBold = ssaOverrideBool(tag.args, sa.SRTBold)
		case "i":
			sa.SRTItalics = ssaOverrideBool(tag.args, sa.SRTItalics)
		case "u":
			sa.SRTUnderline = ssaOverrideBool(tag.args, sa.SRTUnderline)
		case "c", "1c":
			sa.SRTColor = ssaOverrideColor(tag.args, sa.SRTColor)
		}
	}
}

// ssaOverrideAlignment parses an `\an8` argument and converts the legacy `\a` layout
func ssaOverrideAlignment(tag ssaOverrideTag, current byte) byte {
	v, err := strconv.Atoi(ssaRegexpOverrideNumber.FindString(tag.args))
	if err != nil {
		return current
	}
	if tag.name == "a" {
		switch v {
		case 1, 2, 3: // bottom row, same as the numpad one
		case 5, 6, 7: // top row, `\a5` is `\an7`
			v += 2
		case 9, 10, 11: // middle row, `\a9` is `\an4`
			v -= 5
		default:
			return current
		}
	}
	if v < 1 || v > 9 {
		return current
	}
	return byte(v)
}

// ssaOverrideBool parses an `\i1` argument, an invalid one leaving the attribute as it is
func ssaOverrideBool(args string, current bool) bool {
	v, err := strconv.Atoi(ssaRegexpOverrideNumber.FindString(args))
	if err != nil {
		return current
	}
	return v != 0
}

// ssaOverrideColor parses colors being written &Hbbggrr& in hexadecimal or as a plain decimal number
func ssaOverrideColor(args string, current *Color) *Color {
	var m = ssaRegexpOverrideColor.FindStringSubmatch(args)
	if m == nil {
		return current
	}
	var base = 10
	if m[1] != "" {
		base = 16
	}
	c, err := newColorFromSSAString(m[2], base)
	if err != nil {
		return current
	}
	return c
}
