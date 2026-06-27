package presenter

import (
	"bytes"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var md = goldmark.New(goldmark.WithExtensions(extension.GFM))

// RenderMarkdown converts markdown to HTML. Raw HTML in the source is escaped
// (goldmark's default, no WithUnsafe), which keeps user content safe.
func RenderMarkdown(src []byte) string {
	var buf bytes.Buffer
	if err := md.Convert(src, &buf); err != nil {
		return "<pre>" + string(src) + "</pre>"
	}
	return buf.String()
}

var (
	highlightStyle     = styles.Get("github")
	highlightFormatter = chromahtml.New(chromahtml.WithClasses(false), chromahtml.WithLineNumbers(true))
)

// Highlight renders source code as highlighted HTML, choosing a lexer by filename.
func Highlight(code, filename string) string {
	lexer := lexers.Match(filename)
	if lexer == nil {
		lexer = lexers.Analyse(code)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return "<pre>" + code + "</pre>"
	}

	var buf bytes.Buffer
	if err := highlightFormatter.Format(&buf, highlightStyle, iterator); err != nil {
		return "<pre>" + code + "</pre>"
	}
	return buf.String()
}
