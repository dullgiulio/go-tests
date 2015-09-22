package main

import "errors"

// TODO: type brace byte
// TODO: better names for stuff (brace -> depth, r -> c)
// TODO: put type of brace into tree

var (
	openBrackets  = []byte{'(', '[', '{'}
	closeBrackets = []byte{')', ']', '}'}
)

var errNotBracket = errors.New("Not a bracket")

func isByte(r byte, lr []byte) bool {
	for i := range lr {
		if r == lr[i] {
			return true
		}
	}
	return false
}

func isOpen(r byte) bool {
	return isByte(r, openBrackets)
}

func isClose(r byte) bool {
	return isByte(r, closeBrackets)
}

func oppositeBrace(r byte) (byte, error) {
	for i := range openBrackets {
		if r == openBrackets[i] {
			return closeBrackets[i], nil
		}
	}
	return 0, errNotBracket
}

func isSpace(c byte) bool {
	if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
		return true
	}
	return false
}

type brace struct {
	current *list
	c       byte
}

type parser struct {
	pos     int
	str     string
	braces  []brace
	current *list
}

func newParser(s string) *parser {
	return &parser{
		str:     s,
		braces:  make([]brace, 0),
		current: newList(),
	}
}

func (p *parser) skipSpaces() {
	for ; p.pos < len(p.str); p.pos++ {
		if !isSpace(p.str[p.pos]) {
			break
		}
	}
}

func (p *parser) nextWord() (end int, err error) {
	for ; p.pos < len(p.str); p.pos++ {
		if p.str[p.pos] == ',' {
			end = p.pos
			p.pos++
			return end, nil
		}
		if isSpace(p.str[p.pos]) {
			break
		}
		if isClose(p.str[p.pos]) {
			break
		}
	}
	end = p.pos
	p.skipSpaces()
	if isClose(p.str[p.pos]) || p.end() {
		return p.pos, nil
	}
	if p.str[p.pos] != ',' {
		return 0, errors.New("unexpected character, want comma")
	}
	p.pos++
	return end, nil
}

func (p *parser) end() bool {
	return p.pos >= len(p.str)
}

func (p *parser) parse() error {
	for !p.end() {
		p.skipSpaces()
		if isOpen(p.str[p.pos]) {
			cb, err := oppositeBrace(p.str[p.pos])
			if err != nil {
				return err
			}
			p.pos++
			p.braces = append(p.braces, brace{p.current, cb})
			p.current = newList()
			continue
		}
		if isClose(p.str[p.pos]) {
			if len(p.braces) == 0 {
				return errors.New("unmatched closed bracket")
			}
			cur := p.current
			if len(p.braces) > 0 {
				p.current = p.braces[len(p.braces)-1].current
				p.current.add(cur)
			}
			var b brace
			b, p.braces = p.braces[len(p.braces)-1], p.braces[:len(p.braces)-1]
			if b.c != p.str[p.pos] {
				return errors.New("unexpected matching bracket")
			}
			p.pos++
			continue
		}
		begin := p.pos
		end, err := p.nextWord()
		if err != nil {
			return err
		}
		p.current.add(newText(p.str[begin:end]))
	}
	return nil
}
