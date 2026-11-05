package parser

import "regexp"

// fencePattern matches the fence lines that open and close AsciiDoc delimited
// blocks: listing (----), literal (....), passthrough (++++), comment (////),
// example (====), sidebar (****), quote (____) and open (--).
var fencePattern = regexp.MustCompile(`^(-{4,}|\.{4,}|\+{4,}|/{4,}|={4,}|\*{4,}|_{4,}|--)$`)

// blockTracker follows delimited blocks line by line, so that a heading or an
// attribute entry shown inside a block isn't taken for a real one.
type blockTracker struct {
	open []string // fences of the blocks currently open, innermost last
}

// step consumes one line and reports whether it was a fence.
func (b *blockTracker) step(line string) bool {
	if !fencePattern.MatchString(line) {
		return false
	}
	if n := len(b.open); n > 0 && b.open[n-1] == line {
		b.open = b.open[:n-1]
		return true
	}
	if b.verbatim() {
		// Inside a verbatim block, any other fence is content.
		return false
	}
	b.open = append(b.open, line)
	return true
}

// inBlock reports whether the tracker is inside any delimited block.
func (b *blockTracker) inBlock() bool {
	return len(b.open) > 0
}

// verbatim reports whether the tracker is inside a block whose lines asciidoctor
// doesn't parse as AsciiDoc: listing, literal, passthrough or comment.
func (b *blockTracker) verbatim() bool {
	for _, fence := range b.open {
		switch fence[0] {
		case '-':
			if fence != "--" { // "--" is an open block, which holds AsciiDoc
				return true
			}
		case '.', '+', '/':
			return true
		}
	}
	return false
}
