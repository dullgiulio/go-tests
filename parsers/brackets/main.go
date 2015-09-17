package main

import (
	"errors"
	"fmt"
	"log"
)

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
	begin, end int
	c          byte
}

type parser struct {
	pos    int
	str    string
	braces []brace
}

func newParser(s string) *parser {
	return &parser{
		str:    s,
		braces: make([]brace, 0),
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
			p.braces = append(p.braces, brace{p.pos, 0, cb})
			continue
		}
		if isClose(p.str[p.pos]) {
			if len(p.braces) == 0 {
				return errors.New("unmatched closed bracket")
			}
			var b brace
			b, p.braces = p.braces[len(p.braces)-1], p.braces[:len(p.braces)-1]
			if b.c != p.str[p.pos] {
				return errors.New("unexpected matching bracket")
			}
			fmt.Printf("sub: %s\n", p.str[b.begin:p.pos])
			p.pos++
			continue
		}
		begin := p.pos
		end, err := p.nextWord()
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", p.str[begin:end])
	}
	return nil
}

func main() {
	p := newParser("(arg0, arg0.1), (arg1, (arg2.1, arg2.2),  arg3, arg4, (arg5.1,arg5.2,  arg5.3 ) )")
	if err := p.parse(); err != nil {
		log.Fatal(err)
	}
}
