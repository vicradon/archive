package markdown

import (
	"fmt"
	"regexp"
	"strings"

	"nb2slides/internal/util"
)

// RenderMarkdown converts notebook markdown to HTML.
// LaTeX is passed through completely raw so MathJax can render it.
// Supports: headings (ATX + setext), bold, italic, inline code, fenced code,
// bullet/ordered lists, blockquotes, HR, $inline math$, $$block math$$,
// \begin{...}...\end{...} LaTeX environments.
func RenderMarkdown(md string) string {
	lines := strings.Split(md, "\n")
	var out strings.Builder
	inCode := false
	inList := false
	inOList := false
	inMath := false       // inside $$ ... $$
	inLatexEnv := false   // inside \begin{...} ... \end{...}
	var mathBuf strings.Builder

	flush := func() {
		if inList {
			out.WriteString("</ul>")
			inList = false
		}
		if inOList {
			out.WriteString("</ol>")
			inOList = false
		}
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// ── \begin{...} LaTeX environment (pass raw to MathJax as display math)
		if !inCode && !inMath && strings.HasPrefix(trimmed, `\begin{`) {
			flush()
			inLatexEnv = true
			mathBuf.WriteString(line)
			mathBuf.WriteString("\n")
			continue
		}
		if inLatexEnv {
			mathBuf.WriteString(line)
			mathBuf.WriteString("\n")
			if strings.Contains(trimmed, `\end{`) {
				// Emit as MathJax display math: wrap in \[ ... \]
				out.WriteString(`<div class="math-display">$$`)
				out.WriteString(mathBuf.String())
				out.WriteString(`$$</div>`)
				mathBuf.Reset()
				inLatexEnv = false
			}
			continue
		}

		// ── $$ display math block
		if trimmed == "$$" {
			if inMath {
				out.WriteString("$$</div>")
				mathBuf.Reset()
				inMath = false
			} else {
				flush()
				out.WriteString(`<div class="math-display">$$`)
				inMath = true
			}
			continue
		}
		if inMath {
			out.WriteString(line)
			out.WriteString("\n")
			continue
		}

		// ── Fenced code block
		if strings.HasPrefix(line, "```") {
			if inCode {
				out.WriteString("</code></pre>")
				inCode = false
			} else {
				flush()
				out.WriteString(`<pre><code>`)
				inCode = true
			}
			continue
		}
		if inCode {
			out.WriteString(util.EscapeHTML(line) + "\n")
			continue
		}

		// ── Setext headings (look one line ahead for === or ---)
		if i+1 < len(lines) {
			next := strings.TrimSpace(lines[i+1])
			if len(next) > 0 && strings.Trim(next, "=") == "" {
				flush()
				out.WriteString(fmt.Sprintf("<h1>%s</h1>", inlineMarkdown(trimmed)))
				// skip the === line on next iteration by blanking it
				lines[i+1] = ""
				continue
			}
			if len(next) > 0 && strings.Trim(next, "-") == "" && len(next) >= 2 {
				flush()
				out.WriteString(fmt.Sprintf("<h2>%s</h2>", inlineMarkdown(trimmed)))
				lines[i+1] = ""
				continue
			}
		}

		// ── ATX Headings
		if strings.HasPrefix(line, "#### ") {
			flush()
			out.WriteString(fmt.Sprintf("<h4>%s</h4>", inlineMarkdown(line[5:])))
			continue
		}
		if strings.HasPrefix(line, "### ") {
			flush()
			out.WriteString(fmt.Sprintf("<h3>%s</h3>", inlineMarkdown(line[4:])))
			continue
		}
		if strings.HasPrefix(line, "## ") {
			flush()
			out.WriteString(fmt.Sprintf("<h2>%s</h2>", inlineMarkdown(line[3:])))
			continue
		}
		if strings.HasPrefix(line, "# ") {
			flush()
			out.WriteString(fmt.Sprintf("<h1>%s</h1>", inlineMarkdown(line[2:])))
			continue
		}

		// ── HR (only bare ---, not setext which was already handled)
		if trimmed == "---" || trimmed == "***" || trimmed == "===" {
			flush()
			out.WriteString("<hr/>")
			continue
		}

		// ── Blockquote
		if strings.HasPrefix(line, "> ") {
			flush()
			out.WriteString(fmt.Sprintf("<blockquote><p>%s</p></blockquote>", inlineMarkdown(line[2:])))
			continue
		}

		// ── Ordered list
		olRe := regexp.MustCompile(`^\d+\.\s+(.*)`)
		if m := olRe.FindStringSubmatch(line); m != nil {
			if inList {
				out.WriteString("</ul>")
				inList = false
			}
			if !inOList {
				out.WriteString("<ol>")
				inOList = true
			}
			out.WriteString(fmt.Sprintf("<li>%s</li>", inlineMarkdown(m[1])))
			continue
		}

		// ── Unordered list
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			if inOList {
				out.WriteString("</ol>")
				inOList = false
			}
			if !inList {
				out.WriteString("<ul>")
				inList = true
			}
			out.WriteString(fmt.Sprintf("<li>%s</li>", inlineMarkdown(line[2:])))
			continue
		}

		// ── Blank line
		if trimmed == "" {
			flush()
			continue
		}

		// ── Normal paragraph line
		flush()
		out.WriteString(fmt.Sprintf("<p>%s</p>\n", inlineMarkdown(line)))
	}

	flush()
	if inCode {
		out.WriteString("</code></pre>")
	}
	if inMath || inLatexEnv {
		// unclosed — emit what we have
		out.WriteString("$$</div>")
	}

	return out.String()
}

// inlineMarkdown handles inline formatting.
// IMPORTANT: $...$ math is passed through raw (not escaped) for MathJax.
func inlineMarkdown(s string) string {
	// Protect inline math $...$ by passing it through raw.
	// We do this by temporarily replacing with placeholders so other regexes don't touch it.
	type placeholder struct{ key, val string }
	var placeholders []placeholder
	mathRe := regexp.MustCompile(`\$[^$\n]+\$`)
	idx := 0
	s = mathRe.ReplaceAllStringFunc(s, func(m string) string {
		key := fmt.Sprintf("\x00MATH%d\x00", idx)
		idx++
		placeholders = append(placeholders, placeholder{key, m})
		return key
	})

	// Bold **...**
	boldRe := regexp.MustCompile(`\*\*(.+?)\*\*`)
	s = boldRe.ReplaceAllString(s, "<strong>$1</strong>")

	// Italic *...*
	italicRe := regexp.MustCompile(`\*(.+?)\*`)
	s = italicRe.ReplaceAllString(s, "<em>$1</em>")

	// Inline code `...`  — escape HTML inside
	codeRe := regexp.MustCompile("`([^`]+)`")
	s = codeRe.ReplaceAllStringFunc(s, func(m string) string {
		inner := m[1 : len(m)-1]
		return fmt.Sprintf("<code>%s</code>", util.EscapeHTML(inner))
	})

	// Links [text](url)
	linkRe := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	s = linkRe.ReplaceAllString(s, `<a href="$2">$1</a>`)

	// Restore math placeholders raw (no escaping)
	for _, p := range placeholders {
		s = strings.ReplaceAll(s, p.key, p.val)
	}

	return s
}

