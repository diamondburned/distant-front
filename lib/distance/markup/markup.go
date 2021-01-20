// Package markup provides functions to handle Distance chat markup format.
package markup

import (
	"html"
	"regexp"
	"strconv"
	"strings"
)

// tagMatcherRegex matches all regexes. It is copy-pasted from the Distance
// server implementation.
var tagMatcherRegex = regexp.MustCompile(
	`\[(?:[0-9A-F]{6}|\/?b|\/?i|\/?u|\/?s|\/?c|\-|\/?sub|\/?sup|\/?url|url=[^\]]*)\]`)

func tagBody(tag string) string {
	return strings.Trim(tag, "[]")
}

func isHex(body string) bool {
	if len(body) != 6 {
		return false
	}
	_, err := strconv.ParseInt(body, 16, 32)
	return err == nil
}

type tagStack struct {
	list []htmlTag
}

func (stack *tagStack) add(on, tag string) {
	stack.list = append(stack.list, htmlTag{on, tag})
}

func (stack *tagStack) get(body string) htmlTag {
	// LIFO; search backwards.
	for i := len(stack.list) - 1; i >= 0; i-- {
		if tag := stack.list[i]; tag.closeOn == body {
			// pop this out
			stack.list = append(stack.list[:i], stack.list[i+1:]...)
			// return
			return tag
		}
	}
	return htmlTag{}
}

func (stack tagStack) mustFlush() bool { return len(stack.list) > 0 }

func (stack *tagStack) flush(builder *strings.Builder) {
	for i := len(stack.list) - 1; i >= 0; i-- {
		builder.WriteString(stack.list[i].closeTag)
	}
	stack.list = stack.list[:0]
}

type htmlTag struct {
	// closeOn describes the tag to write the close on.
	closeOn string
	// closeTag describes the HTML tag to be written.
	closeTag string
}

// ToHTML converts the given Distance markup string to HTML.
func ToHTML(markup string) string {
	matches := tagMatcherRegex.FindAllStringSubmatchIndex(markup, -1)
	if len(matches) == 0 {
		// The markup has no tags, therefore the whole thing is one big text.
		return html.EscapeString(markup)
	}

	buf := strings.Builder{}
	buf.Grow(len(markup) + 128) // alloc extra for tags

	var prev int
	var stack tagStack
	var htmlTagName string

	for _, match := range matches {
		// write prefix
		buf.WriteString(html.EscapeString(markup[prev:match[0]]))
		// increment previous cursor
		prev = match[1]

		// parse the tag
		tagBody := tagBody(markup[match[0]:match[1]])
		if len(tagBody) == 0 {
			goto writeLiteral
		}

		// handle if this is a closing tag
		if tagBody == "-" || strings.HasPrefix(tagBody, "/") {
			tag := stack.get(tagBody)
			// invalid tag; write as literal.
			if tag.closeOn == "" {
				goto writeLiteral
			}

			buf.WriteString(tag.closeTag)
			continue
		}

		// attempt to parse the color
		if isHex(tagBody) {
			buf.WriteString(`<span style="color:#`)
			buf.WriteString(tagBody)
			buf.WriteString(`">`)

			stack.add("-", "</span>")
			continue
		}

		// attempt to parse URL
		if strings.HasPrefix(tagBody, "url=") {
			buf.WriteString(`<a href="`)
			buf.WriteString(html.EscapeString(strings.TrimPrefix(tagBody, "url=")))
			buf.WriteString(`">`)

			stack.add("/url", "</a>")
			continue
		}

		switch tagBody {
		case "b":
			htmlTagName = "b"
		case "i":
			htmlTagName = "i"
		case "u":
			htmlTagName = "u"
		case "s":
			htmlTagName = "del"
		case "sub":
			htmlTagName = "sub"
		case "sup":
			htmlTagName = "sup"
		case "c":
			stack.add("/c", "")
			continue // unsure what c is
		}

		if htmlTagName != "" {
			buf.WriteByte('<')
			buf.WriteString(htmlTagName)
			buf.WriteByte('>')

			stack.add("/"+tagBody, "</"+htmlTagName+">")
			continue
		}

	writeLiteral:
		buf.WriteString(html.EscapeString(markup[match[0]:match[1]]))
	}

	// flush the rest of the string
	if prev < len(markup) {
		buf.WriteString(markup[prev:])
	}

	// flush the remaining close tags
	if stack.mustFlush() {
		stack.flush(&buf)
	}

	return buf.String()
}
