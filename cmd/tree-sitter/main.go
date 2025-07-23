package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

// XML structures for Chroma lexer definition
type Lexer struct {
	XMLName xml.Name `xml:"lexer"`
	Config  Config   `xml:"config"`
	Rules   Rules    `xml:"rules"`
}

type Config struct {
	Name     string `xml:"name"`
	Alias    string `xml:"alias"`
	Filename string `xml:"filename"`
	MimeType string `xml:"mime_type"`
}

type Rules struct {
	States []State `xml:"state"`
}

type State struct {
	Name  string `xml:"name,attr"`
	Rules []Rule `xml:"rule"`
}

type Rule struct {
	Pattern  string    `xml:"pattern,attr,omitempty"`
	Include  string    `xml:"include,attr,omitempty"`
	Push     string    `xml:"push,attr,omitempty"`
	Pop      int       `xml:"pop,attr,omitempty"`
	Token    *Token    `xml:"token,omitempty"`
	ByGroups *ByGroups `xml:"bygroups,omitempty"`
}

type Token struct {
	Type string `xml:"type,attr"`
}

type ByGroups struct {
	Tokens []Token `xml:"token"`
}

// TokenType represents Chroma token types
type TokenType string

const (
	Keyword               TokenType = "Keyword"
	KeywordConstant       TokenType = "KeywordConstant"
	KeywordNamespace      TokenType = "KeywordNamespace"
	KeywordType           TokenType = "KeywordType"
	NameFunction          TokenType = "NameFunction"
	NameBuiltin           TokenType = "NameBuiltin"
	NameVariable          TokenType = "NameVariable"
	NameAttribute         TokenType = "NameAttribute"
	NameConstant          TokenType = "NameConstant"
	LiteralNumber         TokenType = "LiteralNumber"
	LiteralString         TokenType = "LiteralString"
	LiteralStringDouble   TokenType = "LiteralStringDouble"
	LiteralStringSingle   TokenType = "LiteralStringSingle"
	LiteralStringEscape   TokenType = "LiteralStringEscape"
	LiteralStringInterpol TokenType = "LiteralStringInterpol"
	CommentSingle         TokenType = "CommentSingle"
	CommentHashbang       TokenType = "CommentHashbang"
	Operator              TokenType = "Operator"
	Punctuation           TokenType = "Punctuation"
	PunctuationSpecial    TokenType = "PunctuationSpecial"
	Text                  TokenType = "Text"
	TextWhitespace        TokenType = "TextWhitespace"
)

// Built-in commands extracted from tree-sitter-nu highlights.scm
var builtinCommands = []string{
	"all", "ansi", "any", "append", "ast", "bits", "bytes", "cal", "cd", "char", "clear",
	"collect", "columns", "compact", "complete", "config", "cp", "date", "debug",
	"decode", "default", "detect", "dfr", "drop", "du", "each", "encode", "enumerate",
	"every", "exec", "exit", "explain", "explore", "export-env", "fill", "filter",
	"find", "first", "flatten", "fmt", "format", "from", "generate", "get", "glob",
	"grid", "group", "group-by", "hash", "headers", "histogram", "history", "http",
	"input", "insert", "inspect", "interleave", "into", "is-empty", "is-not-empty",
	"is-terminal", "items", "join", "keybindings", "kill", "last", "length",
	"let-env", "lines", "load-env", "ls", "math", "merge", "metadata", "mkdir",
	"mktemp", "move", "mv", "nu-check", "nu-highlight", "open", "panic", "par-each",
	"parse", "path", "plugin", "port", "prepend", "print", "ps", "query", "random",
	"range", "reduce", "reject", "rename", "reverse", "rm", "roll", "rotate",
	"run-external", "save", "schema", "select", "seq", "shuffle", "skip", "sleep",
	"sort", "sort-by", "split", "split-by", "start", "stor", "str", "sys", "table",
	"take", "tee", "term", "timeit", "to", "touch", "transpose", "tutor", "ulimit",
	"uname", "uniq", "uniq-by", "update", "upsert", "url", "values", "view", "watch",
	"where", "which", "whoami", "window", "with-env", "wrap", "zip",
}

// Keywords extracted from tree-sitter-nu highlights.scm
var keywords = []string{
	"def", "alias", "export-env", "export", "extern", "module",
	"let", "let-env", "mut", "const", "hide-env", "source", "source-env",
	"overlay", "loop", "while", "error", "do", "if", "else", "try", "catch", "match",
	"break", "continue", "return", "hide", "use", "for", "in", "list", "new", "as", "make",
}

// Constants and literals
var constants = []string{
	"true", "false", "null", "nothing",
}

func createRule(pattern string, tokenType TokenType, states ...string) Rule {
	rule := Rule{
		Pattern: pattern,
		Token:   &Token{Type: string(tokenType)},
	}
	if len(states) > 0 {
		if states[0] == "" {
			rule.Pop = 1
		} else {
			rule.Push = states[0]
		}
	}
	return rule
}

func createByGroupsRule(pattern string, tokenTypes ...TokenType) Rule {
	tokens := make([]Token, len(tokenTypes))
	for i, tt := range tokenTypes {
		tokens[i] = Token{Type: string(tt)}
	}
	return Rule{
		Pattern:  pattern,
		ByGroups: &ByGroups{Tokens: tokens},
	}
}

func escapeRegex(s string) string {
	// Escape special regex characters
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ".", "\\.")
	s = strings.ReplaceAll(s, "*", "\\*")
	s = strings.ReplaceAll(s, "+", "\\+")
	s = strings.ReplaceAll(s, "?", "\\?")
	s = strings.ReplaceAll(s, "[", "\\[")
	s = strings.ReplaceAll(s, "]", "\\]")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	s = strings.ReplaceAll(s, "{", "\\{")
	s = strings.ReplaceAll(s, "}", "\\}")
	s = strings.ReplaceAll(s, "^", "\\^")
	s = strings.ReplaceAll(s, "$", "\\$")
	s = strings.ReplaceAll(s, "|", "\\|")
	return s
}

func createWordBoundaryPattern(words []string) string {
	if len(words) == 0 {
		return ""
	}

	// Sort by length (longest first) to avoid partial matches
	sort.Slice(words, func(i, j int) bool {
		return len(words[i]) > len(words[j])
	})

	var escaped []string
	for _, word := range words {
		escaped = append(escaped, escapeRegex(word))
	}

	return `\b(` + strings.Join(escaped, "|") + `)\b`
}

func generateLexer() *Lexer {
	// Root state rules
	rootRules := []Rule{
		{Include: "basic"},
		{Include: "data"},
	}

	// Basic syntax rules
	basicRules := []Rule{
		// Shebang
		createRule(`\A#!.+\n`, CommentHashbang),

		// Comments
		createRule(`#.*\n`, CommentSingle),

		// Keywords
		createByGroupsRule(`(`+createWordBoundaryPattern(keywords)+`)(\s*)`, Keyword, TextWhitespace),

		// Built-in commands
		createByGroupsRule(`(`+createWordBoundaryPattern(builtinCommands)+`)(\s*)`, NameBuiltin, TextWhitespace),

		// Constants
		createByGroupsRule(`(`+createWordBoundaryPattern(constants)+`)(\s*)`, KeywordConstant, TextWhitespace),

		// External commands (starting with ^)
		createByGroupsRule(`(\^)(\w+)`, Operator, NameFunction),

		// Function definitions
		createByGroupsRule(`(def|alias)(\s+)(\w+)`, Keyword, TextWhitespace, NameFunction),

		// Variable assignments
		createByGroupsRule(`(\$?\w+)(\s*)(=|\+=|-=|\*=|/=|\+\+=)`, NameVariable, TextWhitespace, Operator),

		// Variables
		createRule(`\$\w+`, NameVariable),
		createRule(`\$&#34;`, LiteralStringDouble, "interpolated_string"),

		// Flags
		createRule(`--\w+(-\w+)*`, NameAttribute),
		createRule(`-\w`, NameAttribute),

		// Operators
		createRule(`==|!=|<=|>=|<|>`, Operator),
		createRule(`\+|-|\*|/|%|\*\*`, Operator),
		createRule(`and|or|not|in`, Operator),
		createRule(`=~|!~|like|not-like`, Operator),
		createRule(`&&|\|\|`, Operator),

		// Redirection operators
		createRule(`o>|out>|e>|err>|e\+o>|err\+out>|o\+e>|out\+err>`, Operator),
		createRule(`o>>|out>>|e>>|err>>|e\+o>>|err\+out>>|o\+e>>|out\+err>>`, Operator),
		createRule(`e>\||err>\||e\+o>\||err\+out>\||o\+e>\||out\+err>\|`, Operator),

		// Range operators
		createRule(`\.\.=?|\.\.<?`, Operator),

		// Pipe and other operators
		createRule(`\|`, Operator),
		createRule(`=>`, Operator),

		// Brackets and punctuation
		createRule(`[\[\]{}()]`, Punctuation),
		createRule(`[,;]`, PunctuationSpecial),
		createRule(`\.\.\.`, PunctuationSpecial),
		createRule(`[@:]`, PunctuationSpecial),
		createRule(`->`, PunctuationSpecial),
	}

	// Data and literals
	dataRules := []Rule{
		// Numbers (various formats)
		createRule(`0x[0-9a-fA-F]+`, LiteralNumber),
		createRule(`0b[01]+`, LiteralNumber),
		createRule(`0o[0-7]+`, LiteralNumber),
		createRule(`\d+(\.\d+)?([eE][+-]?\d+)?`, LiteralNumber),

		// Durations and filesizes
		createRule(`\d+(\.\d+)?(ns|us|ms|sec|min|hr|day|wk)`, LiteralNumber),
		createRule(`\d+(\.\d+)?(B|KB|MB|GB|TB|PB|EB|ZB|YB|KiB|MiB|GiB|TiB|PiB|EiB|ZiB|YiB)`, LiteralNumber),

		// Dates
		createRule(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})`, LiteralNumber),
		createRule(`\d{4}-\d{2}-\d{2}`, LiteralNumber),

		// Raw strings
		createRule(`r#&#34;[^&#34;]*&#34;#`, LiteralString),
		createRule(`r&#39;[^&#39;]*&#39;`, LiteralString),

		// Regular strings
		createRule(`&#34;([^&#34;\\]|\\.)*&#34;`, LiteralStringDouble),
		createRule(`&#39;([^&#39;\\]|\\.)*&#39;`, LiteralStringSingle),

		// Escape sequences
		createRule(`\\[\\&#39;&#34;nrt0$]`, LiteralStringEscape),
		createRule(`\\x[0-9a-fA-F]{2}`, LiteralStringEscape),
		createRule(`\\u\{[0-9a-fA-F]+\}`, LiteralStringEscape),

		// Whitespace
		createRule(`\s+`, TextWhitespace),

		// Default text
		createRule(`[^\s\[\]{}()$&#34;&#39;`+"`"+`\\<>&|;#]+`, Text),
	}

	// Interpolated string state
	interpolatedStringRules := []Rule{
		createRule(`&#34;`, LiteralStringDouble, ""), // Pop state
		createRule(`\(`, LiteralStringInterpol, "interpolation"),
		createRule(`([^&#34;\\(]|\\.)+`, LiteralStringDouble),
	}

	// Interpolation state
	interpolationRules := []Rule{
		createRule(`\)`, LiteralStringInterpol, ""), // Pop state
		{Include: "root"},
	}

	return &Lexer{
		Config: Config{
			Name:     "Nu",
			Alias:    "nu",
			Filename: "*.nu",
			MimeType: "text/plain",
		},
		Rules: Rules{
			States: []State{
				{Name: "root", Rules: rootRules},
				{Name: "basic", Rules: basicRules},
				{Name: "data", Rules: dataRules},
				{Name: "interpolated_string", Rules: interpolatedStringRules},
				{Name: "interpolation", Rules: interpolationRules},
			},
		},
	}
}

func main() {
	// Create the lexer
	lexer := generateLexer()

	// Marshal to XML with proper formatting
	output, err := xml.MarshalIndent(lexer, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling XML: %v", err)
	}

	// Add XML declaration
	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>` + "\n" + string(output)

	// Clean up the XML formatting for better readability
	xmlContent = strings.ReplaceAll(xmlContent, `&gt;`, `>`)
	xmlContent = strings.ReplaceAll(xmlContent, `&lt;`, `<`)
	xmlContent = strings.ReplaceAll(xmlContent, `&amp;`, `&`)
	// Keep quotes escaped in XML attributes
	// xmlContent = strings.ReplaceAll(xmlContent, `&#39;`, `'`)
	// xmlContent = strings.ReplaceAll(xmlContent, `&#34;`, `"`)

	// Write to file
	outputPath := "../../lexers/embedded/nu.xml"
	err = os.WriteFile(outputPath, []byte(xmlContent), 0644)
	if err != nil {
		log.Fatalf("Error writing file: %v", err)
	}

	fmt.Printf("Generated new nu.xml lexer at: %s\n", outputPath)
	fmt.Println("\nThe lexer includes comprehensive support for:")
	fmt.Println("- Keywords and built-in commands")
	fmt.Println("- Variables and assignments")
	fmt.Println("- String interpolation")
	fmt.Println("- Numbers, dates, durations, and filesizes")
	fmt.Println("- Comments and operators")
	fmt.Println("- Flags and punctuation")
	fmt.Println("\nThis lexer was generated based on the official tree-sitter-nu grammar")
	fmt.Println("and should provide much better syntax highlighting than the previous version.")
}
