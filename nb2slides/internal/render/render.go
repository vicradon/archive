package render

import (
	"fmt"
	"html"
	"strings"

	"nb2slides/internal/util"
)

// Theme controls the colour scheme applied to the rendered HTML.
// Each value maps to a body CSS class (theme-white, theme-dark, theme-sepia).
type Theme string

const (
	ThemeWhite Theme = "white"
	ThemeDark  Theme = "dark"
	ThemeSepia Theme = "sepia"
)

// ── SVG background patterns ────────────────────────────────────────────────
//
// Each pattern is a tiny, tileable SVG encoded as a data URI.
// They are intentionally low-contrast so they sit behind content without
// fighting it.  The fill colour is kept neutral (currentColor / fixed hex)
// so the theme layer can tint them via CSS filter if needed.

const svgDots = `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='20' height='20'%3E%3Ccircle cx='2' cy='2' r='1.2' fill='%23c8c8c8' fill-opacity='.55'/%3E%3C/svg%3E")`

const svgGrid = `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='24' height='24'%3E%3Cpath d='M 24 0 L 0 0 0 24' fill='none' stroke='%23c0c0c0' stroke-width='.6' stroke-opacity='.5'/%3E%3C/svg%3E")`

const svgWaves = `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='60' height='12'%3E%3Cpath d='M0 6 Q15 0 30 6 Q45 12 60 6' fill='none' stroke='%23c0c0c0' stroke-width='.8' stroke-opacity='.45'/%3E%3C/svg%3E")`

const svgCrosshatch = `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16'%3E%3Cpath d='M0 0 L16 16 M16 0 L0 16' stroke='%23c0c0c0' stroke-width='.5' stroke-opacity='.4'/%3E%3C/svg%3E")`

// ── Base CSS (layout, typography, print) ──────────────────────────────────

const baseCSS = `
* { box-sizing: border-box; margin: 0; padding: 0; }

body {
  font-family: Georgia, 'Times New Roman', serif;
  font-size: 13px;
  color: #111;
}

/* ── Print: strip chrome, keep layout ── */
@media print {
  body { background: #fff !important; margin: 0; padding: 0; }
  .slide {
    width: 100%;
    padding: 14mm 18mm 13mm;
    page-break-after: always;
    break-after: page;
    position: relative;
    border: none !important;
    box-shadow: none !important;
    background-image: none !important;
  }
  .export-bar { display: none !important; }
}

/* ── Slide wrapper ── */
@media screen {
  body { padding: 28px 20px; }
  .slide {
    width: 210mm;
    min-height: 148mm;
    margin: 0 auto 20px;
    border-radius: 4px;
    padding: 14mm 18mm 13mm;
    page-break-after: always;
    position: relative;
    background-repeat: repeat;
    background-size: auto;
  }
}

/* ── Slide label ── */
.slide-label {
  position: absolute;
  top: 7mm;
  right: 10mm;
  font-size: 9.5px;
  font-family: 'Helvetica Neue', Arial, sans-serif;
  letter-spacing: 0.4px;
}

/* ── Markdown cells ── */
.cell-md {
  font-family: Georgia, 'Times New Roman', serif;
  font-size: 13.5px;
  line-height: 1.65;
}
.cell-md h1 {
  font-size: 20px;
  font-weight: 700;
  padding-bottom: 5px;
  margin-bottom: 12px;
  font-family: 'Helvetica Neue', Arial, sans-serif;
}
.cell-md h2 {
  font-size: 16px;
  font-weight: 600;
  padding-bottom: 4px;
  margin-bottom: 9px;
  font-family: 'Helvetica Neue', Arial, sans-serif;
}
.cell-md h3 {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 6px;
  font-family: 'Helvetica Neue', Arial, sans-serif;
}
.cell-md p  { margin-bottom: 7px; }
.cell-md ul, .cell-md ol { margin-left: 22px; margin-bottom: 7px; }
.cell-md li { margin-bottom: 3px; }
.cell-md code {
  padding: 1px 4px;
  border-radius: 3px;
  font-size: 11.5px;
  font-family: 'Courier New', Courier, monospace;
}
.cell-md pre {
  padding: 8px 10px;
  border-radius: 3px;
  margin-bottom: 8px;
  overflow-x: auto;
}
.cell-md pre code { border: none; background: none; padding: 0; }
.cell-md blockquote {
  border-left: 3px solid;
  padding-left: 10px;
  margin-bottom: 7px;
}
.cell-md .MJX-TEX { font-size: 13px; }
.mjx-chtml { margin: 6px 0 !important; }
.math-display { margin: 8px 0; overflow-x: auto; }

/* ── Code cells with line numbers ── */
.cell-code {
  border-left: 3px solid;
  border-radius: 0 3px 3px 0;
}
.cell-code pre {
  padding: 9px 12px 9px 0;
  overflow-x: auto;
  white-space: pre;
  counter-reset: ln;
}
.cell-code code {
  font-family: 'Courier New', Courier, monospace;
  font-size: 12px;
  line-height: 1.6;
  display: block;
}

/* Each logical line is wrapped in <span class="ln"> by Go */
.cell-code .ln {
  display: block;
  padding-left: 3.2em;
  position: relative;
  white-space: pre-wrap;
  word-break: break-all;
}
.cell-code .ln::before {
  counter-increment: ln;
  content: counter(ln);
  position: absolute;
  left: 0;
  width: 2.4em;
  text-align: right;
  font-size: 10px;
  line-height: 1.6;
  padding-right: 0.5em;
  user-select: none;
  opacity: 0.45;
}

/* ── Output area ── */
.cell-output {
  margin-top: 9px;
  border-top: 1px dashed;
  padding-top: 7px;
}
.output-label {
  font-size: 9.5px;
  font-family: 'Helvetica Neue', Arial, sans-serif;
  text-transform: uppercase;
  letter-spacing: 0.6px;
  margin-bottom: 4px;
  opacity: 0.5;
}
.out-stdout {
  border-radius: 3px;
  padding: 5px 8px;
  font-family: 'Courier New', Courier, monospace;
  font-size: 11.5px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
}
.out-stderr {
  border-left: 3px solid #d44;
  padding: 5px 8px;
  font-family: 'Courier New', Courier, monospace;
  font-size: 11.5px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
  color: #c00;
}
.out-image img {
  max-width: 100%;
  max-height: 220px;
  display: block;
  margin: 4px 0;
}
.out-html { font-size: 11.5px; overflow-x: auto; }
.out-html table { border-collapse: collapse; font-size: 11px; }
.out-html td, .out-html th { border: 1px solid; padding: 3px 7px; }
.out-html th { font-weight: 600; }

/* ── Raw cells ── */
.cell-raw pre {
  padding: 8px 10px;
  border-radius: 3px;
  font-family: 'Courier New', Courier, monospace;
  font-size: 11.5px;
  white-space: pre-wrap;
}

/* ── Export / theme bar ── */
.export-bar {
  position: fixed;
  top: 14px;
  right: 18px;
  z-index: 100;
  display: flex;
  gap: 6px;
  align-items: center;
}
.export-bar button {
  padding: 5px 13px;
  font-size: 12px;
  border-radius: 4px;
  cursor: pointer;
  font-family: 'Helvetica Neue', Arial, sans-serif;
  border: 1px solid transparent;
  transition: opacity .15s;
}
.export-bar button:hover { opacity: 0.8; }
`

// ── Theme CSS blocks ──────────────────────────────────────────────────────
//
// Convention: every theme sets CSS custom properties on body.theme-*,
// then all colour-bearing rules reference those properties.
// This means base rules above stay colour-agnostic; only these blocks
// carry colour opinions.

const themeWhite = `
/* ════ THEME: WHITE ════ */
body.theme-white {
  --page-bg:       #f0f0f0;
  --slide-bg:      #ffffff;
  --slide-border:  #d8d8d8;
  --slide-shadow:  0 1px 4px rgba(0,0,0,.09);
  --slide-pattern: ` + svgDots + `;
  --text:          #1a1a1a;
  --text-muted:    #bbb;
  --heading-rule:  #555;
  --h2-rule:       #ccc;
  --blockquote-border: #ccc;
  --code-bg:       #f5f5f5;
  --code-border:   #e0e0e0;
  --code-text:     #1a1a1a;
  --code-accent:   #5b8dd9;
  --output-dashed: #ddd;
  --stdout-bg:     #f7f7f7;
  --raw-bg:        #fffdf0;
  --raw-border:    #e8d44d;
  --btn-bg:        #fff;
  --btn-border:    #999;
  --btn-text:      #222;
  --table-border:  #ddd;
  --th-bg:         #f3f3f3;
}

@media screen {
  body.theme-white              { background: var(--page-bg); }
  body.theme-white .slide       {
    background-color: var(--slide-bg);
    background-image: var(--slide-pattern);
    border: 1px solid var(--slide-border);
    box-shadow: var(--slide-shadow);
  }
}
`

const themeDark = `
/* ════ THEME: DARK ════ */
body.theme-dark {
  --page-bg:       #18181b;
  --slide-bg:      #1e1e24;
  --slide-border:  #2e2e38;
  --slide-shadow:  0 2px 8px rgba(0,0,0,.45);
  --slide-pattern: ` + svgGrid + `;
  --text:          #e2e2e8;
  --text-muted:    #555;
  --heading-rule:  #666;
  --h2-rule:       #3a3a48;
  --blockquote-border: #444;
  --code-bg:       #13131a;
  --code-border:   #2a2a38;
  --code-text:     #c8d0e0;
  --code-accent:   #7aa2f7;
  --output-dashed: #2e2e3e;
  --stdout-bg:     #111118;
  --raw-bg:        #1a1810;
  --raw-border:    #7a6a18;
  --btn-bg:        #26262e;
  --btn-border:    #44445a;
  --btn-text:      #c8c8d8;
  --table-border:  #2e2e40;
  --th-bg:         #252530;
}

@media screen {
  body.theme-dark              { background: var(--page-bg); color: var(--text); }
  body.theme-dark .slide       {
    background-color: var(--slide-bg);
    background-image: var(--slide-pattern);
    border: 1px solid var(--slide-border);
    box-shadow: var(--slide-shadow);
  }
}
`

const themeSepia = `
/* ════ THEME: SEPIA ════ */
body.theme-sepia {
  --page-bg:       #e8e0d0;
  --slide-bg:      #faf6ee;
  --slide-border:  #d4c8ae;
  --slide-shadow:  0 1px 5px rgba(100,80,40,.13);
  --slide-pattern: ` + svgWaves + `;
  --text:          #2c2215;
  --text-muted:    #a89880;
  --heading-rule:  #8a6a40;
  --h2-rule:       #c8b890;
  --blockquote-border: #c0a878;
  --code-bg:       #f0e8d8;
  --code-border:   #d8c8a8;
  --code-text:     #2c2215;
  --code-accent:   #a05820;
  --output-dashed: #c8b890;
  --stdout-bg:     #ede5d5;
  --raw-bg:        #fdfae8;
  --raw-border:    #c8a832;
  --btn-bg:        #f5efe0;
  --btn-border:    #b8a880;
  --btn-text:      #3a2a10;
  --table-border:  #c8b890;
  --th-bg:         #ede0c8;
}

@media screen {
  body.theme-sepia              { background: var(--page-bg); }
  body.theme-sepia .slide       {
    background-color: var(--slide-bg);
    background-image: var(--slide-pattern);
    border: 1px solid var(--slide-border);
    box-shadow: var(--slide-shadow);
  }
}
`

// ── Variable-driven overrides (applied to all themes via CSS vars) ─────────
//
// Rather than repeating every rule three times, base rules use var() and
// we only override the variables per theme above.  This block wires the
// var()-bearing rules that need colour.

const varOverrides = `
.slide-label          { color: var(--text-muted, #bbb); }

.cell-md              { color: var(--text, #1a1a1a); }
.cell-md h1           { border-bottom: 1.5px solid var(--heading-rule, #555); }
.cell-md h2           { border-bottom: 1px solid var(--h2-rule, #ccc); }
.cell-md code         { background: var(--code-bg, #f5f5f5); border: 1px solid var(--code-border, #e0e0e0); }
.cell-md pre          { background: var(--code-bg, #f5f5f5); border: 1px solid var(--code-border, #e0e0e0); }
.cell-md blockquote   { border-left-color: var(--blockquote-border, #ccc); color: var(--text-muted, #555); }

.cell-code            { background: var(--code-bg, #fafafa); border-left-color: var(--code-accent, #5b8dd9); }
.cell-code code       { color: var(--code-text, #1a1a1a); }

.cell-output          { border-top-color: var(--output-dashed, #ddd); }
.out-stdout           { background: var(--stdout-bg, #f7f7f7); color: var(--text, #222); }
.out-html td, .out-html th { border-color: var(--table-border, #ddd); }
.out-html th          { background: var(--th-bg, #f3f3f3); }

.cell-raw pre         { background: var(--raw-bg, #fffdf0); border: 1px solid var(--raw-border, #e8d44d); }

.export-bar button    { background: var(--btn-bg, #fff); border-color: var(--btn-border, #888); color: var(--btn-text, #222); }
`

// ── Theme toggle script ────────────────────────────────────────────────────

const themeToggleScript = `<script>
(function(){
  var themes = ['theme-white','theme-dark','theme-sepia'];
  var labels = {
    'theme-white': 'White',
    'theme-dark':  'Dark',
    'theme-sepia': 'Sepia'
  };
  function current(){
    for(var i=0;i<themes.length;i++){
      if(document.body.classList.contains(themes[i])) return i;
    }
    return 0;
  }
  window._cycleTheme = function(){
    var i = current();
    var next = themes[(i+1)%themes.length];
    document.body.classList.remove(themes[i]);
    document.body.classList.add(next);
    var btn = document.getElementById('btn-theme');
    if(btn) btn.textContent = labels[next];
    try{ localStorage.setItem('nb2slides-theme', next); }catch(e){}
  };
  // restore saved theme on load
  try{
    var saved = localStorage.getItem('nb2slides-theme');
    if(saved && themes.indexOf(saved)>-1){
      document.body.classList.remove(themes[0],themes[1],themes[2]);
      document.body.classList.add(saved);
    }
  }catch(e){}
})();
</script>`

// MathJaxScript is included once in the <head> to render LaTeX in markdown cells.
const MathJaxScript = `<script>
window.MathJax = {
  tex: {
    inlineMath: [['$', '$']],
    displayMath: [['$$', '$$'], ['\\[', '\\]']],
    processEscapes: true,
    packages: {'[+]': ['ams']}
  },
  options: { skipHtmlTags: ['script','noscript','style','textarea','pre','code'] },
  startup: { typeset: true }
};
</script>
<script defer src="https://cdn.jsdelivr.net/npm/mathjax@3/es5/tex-chtml.js"></script>`

// ── Public API ─────────────────────────────────────────────────────────────

// HTMLHead returns a complete <head> block including CSS for all three themes.
//
// theme    – initial theme; defaults to ThemeWhite if empty.
// extraCSS – caller-supplied overrides appended after the built-in styles.
func HTMLHead(title string, theme Theme, extraCSS string) string {
	if theme == "" {
		theme = ThemeWhite
	}

	var b strings.Builder
	b.WriteString(`<!DOCTYPE html>`)
	b.WriteString(fmt.Sprintf(`<html lang="en" class=""><head>`))
	b.WriteString(`<meta charset="UTF-8">`)
	b.WriteString(`<meta name="viewport" content="width=device-width,initial-scale=1">`)
	b.WriteString(fmt.Sprintf(`<title>%s</title>`, util.EscapeHTML(title)))

	b.WriteString(`<style>`)
	b.WriteString(baseCSS)
	b.WriteString(themeWhite)
	b.WriteString(themeDark)
	b.WriteString(themeSepia)
	b.WriteString(varOverrides)
	if extraCSS != "" {
		b.WriteString(extraCSS)
	}
	b.WriteString(`</style>`)

	b.WriteString(MathJaxScript)
	b.WriteString(themeToggleScript)
	b.WriteString(`</head>`)

	// Open body with initial theme class
	b.WriteString(fmt.Sprintf(`<body class="theme-%s">`, theme))

	return b.String()
}

// ExportBar returns the fixed UI bar rendered at the top-right of the page.
// It contains a Print button and a Theme toggle.  Call this right after
// opening <body> (HTMLHead already writes the <body> open tag).
func ExportBar(theme Theme) string {
	if theme == "" {
		theme = ThemeWhite
	}
	label := map[Theme]string{
		ThemeWhite: "White",
		ThemeDark:  "Dark",
		ThemeSepia: "Sepia",
	}[theme]
	return fmt.Sprintf(
		`<div class="export-bar">`+
			`<button id="btn-theme" onclick="window._cycleTheme()">%s</button>`+
			`<button onclick="window.print()">Print / PDF</button>`+
			`</div>`,
		label,
	)
}

// LineNumberedCode wraps each line of src in a <span class="ln"> so that the
// CSS counter can render gutter line numbers without any runtime JS.
// src should be the raw (already HTML-escaped) source text of the cell.
//
// Usage:
//
//	fmt.Fprintf(w, `<div class="cell-code"><pre><code>%s</code></pre></div>`,
//	    render.LineNumberedCode(html.EscapeString(sourceText)))
func LineNumberedCode(src string) string {
	lines := strings.Split(src, "\n")
	// Trim a trailing empty line produced by a final newline
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	var b strings.Builder
	b.Grow(len(src) + len(lines)*32)
	for _, line := range lines {
		b.WriteString(`<span class="ln">`)
		b.WriteString(html.EscapeString(line))
		b.WriteString("</span>\n")
	}
	return b.String()
}
