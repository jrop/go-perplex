package perplex

import (
	"testing"
)

func makeScanner(src string) (Lexer, Scanner) {
	l := NewLexer()
	l.Define("WHITESPACE", `\s+`, true)
	l.Define("COMMENT", "//[^\r\n]*", true)

	l.Define("ID", `[a-zA-Z_$][a-zA-Z0-9_$]*`, false)

	// numbers:
	l.Define("HEX_NUMBER", `0[xX][0-9a-fA-F]+`, false)
	l.Define("OCT_NUMBER", `0[0-7]+`, false)
	l.Define("BIN_NUMBER", `0[bB][01]+`, false)
	// now for decimals, integers, and scientific notation:
	l.Define("NUMBER", `[0-9]+(\.[0-9]+, false)?[eE][+-]?[0-9]+`, false)
	l.Define("NUMBER", `[0-9]+\.[0-9]+`, false)
	l.Define("NUMBER", `[0-9]+`, false)
	return l, l.CreateScanner(src)
}

func TestLexer(t *testing.T) {
	_, s := makeScanner("foo bar baz")

	// foo
	tkn, err := s.Expect("ID")
	if err != nil {
		t.Error(err)
	}
	if tkn.Text() != "foo" {
		t.Errorf("Expected 'foo', got '%s'", tkn.Text())
	}
	if tkn.Pos() != 0 {
		t.Errorf("Expected token position to be 0, got %d", tkn.Pos())
	}
	if tkn.End() != 3 {
		t.Errorf("Expected token end to be 3, got %d", tkn.End())
	}

	// bar
	tkn, err = s.Expect("ID")
	if err != nil {
		t.Error(err)
	}
	if tkn.Text() != "bar" {
		t.Errorf("Expected 'foo', got '%s'", tkn.Text())
	}
	if len(tkn.SkippedTokens()) != 1 {
		t.Errorf("Expected 1 skipped token")
	}
	if tkn.SkippedTokens()[0].Kind() != "WHITESPACE" {
		t.Errorf("Expected skipped token to be whitespace")
	}
	if tkn.Pos() != 4 {
		t.Errorf("Expected token position to be 4, got %d", tkn.Pos())
	}
	if tkn.End() != 7 {
		t.Errorf("Expected token end to be 7, got %d", tkn.End())
	}

	// baz:
	tkn, err = s.Expect("ID")
	if err != nil {
		t.Error(err)
	}
	if tkn.Text() != "baz" {
		t.Errorf("Expected 'foo', got '%s'", tkn.Text())
	}
	if len(tkn.SkippedTokens()) != 1 {
		t.Errorf("Expected 1 skipped token")
	}
	if tkn.SkippedTokens()[0].Kind() != "WHITESPACE" {
		t.Errorf("Expected skipped token to be whitespace")
	}
}

func TestUnexpected(t *testing.T) {
	src := "foo {{{ bar baz"
	l, _ := makeScanner(src)
	tkn := l.ReadToken(src, 4)
	if tkn.Kind() != "UNEXPECTED" {
		t.Errorf("Expected 'UNEXPECTED', got '%s'", tkn.Kind())
	}
	if tkn.Text() != "{{{" {
		t.Errorf("Expected '{{{', got '%s'", tkn.Text())
	}
}

func TestKeyword(t *testing.T) {
	l := NewLexer()
	l.DefineKeyword("function")
	l.Define("WHITESPACE", `\s+`, true)
	l.Define("ID", `[a-zA-Z_$][a-zA-Z0-9_$]*`, false)

	s := l.CreateScanner("function foo")
	tkn, err := s.Expect("function")
	if err != nil {
		t.Error(err)
	}
	if tkn.Kind() != "function" {
		t.Errorf("Expected 'function', got '%s'", tkn.Kind())
	}

	t.Logf("the rest after scanning `function`: %s\n", s.src[s.pos:])

	tkn, err = s.Expect("ID")
	if err != nil {
		t.Error(err)
	}
	if tkn.Kind() != "ID" {
		t.Errorf("Expected 'ID', got '%s'", tkn.Kind())
	}
	if tkn.Text() != "foo" {
		t.Errorf("Expected 'foo', got '%s'", tkn.Text())
	}
}

func TestOperator(t *testing.T) {
	l := NewLexer()
	l.DefineOperator("=")

	tkn := l.ReadToken("foo = bar", 4)
	if tkn.Kind() != "=" {
		t.Errorf("Expected '=', got '%s'", tkn.Kind())
	}
}
