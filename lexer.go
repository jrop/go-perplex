package perplex

import (
	"fmt"
	"regexp"
)

// Token {{{
type tokenData struct {
	kind          string
	text          string
	pos           uint
	skip          bool
	skippedTokens []Token
}

type Token struct {
	*tokenData
}

func NewToken(kind string, text string, pos uint, skip bool) Token {
	return Token{
		&tokenData{
			kind: kind,
			text: text,
			pos:  pos,
			skip: skip,
		},
	}
}

func (t Token) SetSkippedTokens(tokens []Token) {
	t.skippedTokens = tokens
}

func (t Token) IsEOF() bool {
	return t.kind == "EOF"
}

func (t Token) IsUnexpected() bool {
	return t.kind == "UNEXPECTED"
}

func (t Token) Skip() bool {
	return t.skip
}

func (t Token) SetSkip(skip bool) {
	t.skip = skip
}

func (t Token) Kind() string {
	return t.kind
}

func (t Token) Text() string {
	return t.text
}

func (t Token) Pos() uint {
	return t.pos
}

func (t Token) End() uint {
	return t.pos + uint(len(t.text))
}

func (t Token) SkippedTokens() []Token {
	return t.skippedTokens
}

// }}}

// TokenDefinition {{{
type TokenDefinition struct {
	regex *regexp.Regexp
	kind  string
	skip  bool
}

func NewTokenDefinition(kind string, regex string, skip bool) TokenDefinition {
	if len(regex) == 0 {
		panic(fmt.Sprintf("Invalid regex for token '%s': empty string", kind))
	}

	if regex[0] != '^' {
		regex = "^" + regex
	}

	return TokenDefinition{
		regex: regexp.MustCompile(regex),
		kind:  kind,
		skip:  skip,
	}
}

// }}}

// Lexer {{{
type Lexer struct {
	tokenDefinitions []TokenDefinition
}

func NewLexer() Lexer {
	return Lexer{
		tokenDefinitions: make([]TokenDefinition, 0),
	}
}

func (l *Lexer) Define(kind string, regex string, skip bool) {
	l.tokenDefinitions = append(l.tokenDefinitions, NewTokenDefinition(kind, regex, skip))
}

func (l *Lexer) DefineKeyword(kind string) {
	l.Define(kind, fmt.Sprintf(`%s\b`, kind), false)
}

func (l *Lexer) DefineOperator(kind string) {
	// first let's escape any special characters in the regex:
	newRe := regexp.QuoteMeta(kind)
	l.Define(kind, newRe, false)
}

func (l Lexer) ReadTokenAt(src string, pos uint) Token {
	// first check if we're at the end of the string
	if pos >= uint(len(src)) {
		return NewToken("EOF", "", pos, false)
	}

	for _, tokenDefinition := range l.tokenDefinitions {
		regex := tokenDefinition.regex
		match := regex.FindStringSubmatch((src)[pos:])
		if match != nil {
			return NewToken(tokenDefinition.kind, match[0], pos, tokenDefinition.skip)
		}
	}

	// Okay, so if we get to this point, we have unexpected input in the string,
	// let's skip ahead and try again, accumulating unexpected charachters until
	// a valid token is found. At that point we can return the unexpected token
	// with a kind of "UNEXPECTED"
	unexpectedStart := pos
	mintUnexpected := func() Token {
		return NewToken("UNEXPECTED", (src)[unexpectedStart:pos], unexpectedStart, false)
	}

	for {
		if pos >= uint(len(src)) {
			// We've reached the end of the string, so return the unexpected token
			return mintUnexpected()
		}

		pos += 1
		next := l.ReadTokenAt(src, pos)
		if next.Kind() != "UNEXPECTED" || next.IsEOF() {
			return mintUnexpected()
		}
		// If we get to this point, we've found another unexpected token, so we
		// need to keep going...
	}
}

func (l Lexer) ReadToken(src string, pos uint) Token {
	skippedTokens := []Token{}
	for {
		tkn := l.ReadTokenAt(src, pos)
		if tkn.Skip() {
			skippedTokens = append(skippedTokens, tkn)
			pos = tkn.End()
			continue
		}
		tkn.SetSkippedTokens(skippedTokens)
		return tkn
	}
}

func (l *Lexer) CreateScanner(src string) Scanner {
	return Scanner{
		&scannerData{
			src:         src,
			pos:         0,
			definitions: l,
		},
	}
}

// }}}

// Scanner {{{
type scannerData struct {
	src         string
	pos         uint
	definitions *Lexer
}
type Scanner struct {
	*scannerData
}

func (l Scanner) Peek() Token {
	return l.definitions.ReadToken(l.src, l.pos)
}

func (l Scanner) Next() Token {
	tkn := l.Peek()
	l.pos = tkn.End()
	return tkn
}

func (l Scanner) Expect(kind string) (Token, error) {
	tkn := l.Next()
	if tkn.Kind() != kind {
		return Token{}, fmt.Errorf("Expected token of kind %s, got %s", kind, tkn.Kind())
	}
	return tkn, nil
}

func LexerIfNext[T any](lex Scanner, kind string, consequent func() T, alternate func() T) T {
	t := lex.Peek()
	if t.Kind() == kind {
		lex.MarkRead(t)
		return consequent()
	} else {
		return alternate()
	}
}

func (l Scanner) Pos() uint {
	return l.pos
}

func (l Scanner) MarkRead(t Token) {
	l.pos = t.End()
}

func (l Scanner) MarkUnread(t Token) {
	l.pos = t.Pos()
}

func (l Scanner) IsEOF() bool {
	return l.pos >= uint(len(l.src))
}

// }}}
