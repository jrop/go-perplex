package perplex

import (
	"testing"
)

func makeScanner(src string) Scanner {
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
	return l.CreateScanner(src)
}

func TestLexer(t *testing.T) {
	l := makeScanner("foo bar baz")

	// foo
	tkn, err := l.Expect("ID")
	if err != nil {
		t.Error(err)
	}
	if tkn.Text() != "foo" {
		t.Errorf("Expected 'foo', got '%s'", tkn.Text())
	}

	// bar
	tkn, err = l.Expect("ID")
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

	// baz:
	tkn, err = l.Expect("ID")
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
