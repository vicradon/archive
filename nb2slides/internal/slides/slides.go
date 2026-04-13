package slides

import (
	"fmt"
	"strings"

	"nb2slides/internal/markdown"
	"nb2slides/internal/notebook"
	"nb2slides/internal/util"
)

// Slide represents one PDF page / one HTML section
type Slide struct {
	ID       string // e.g. "1", "5-a", "5-b"
	CellIdx  int    // original cell index (1-based)
	CellType string
	HTML     string // rendered HTML content for this slide
}

// BuildSlides builds slides from a notebook
func BuildSlides(nb *notebook.Notebook, maxLines int, showOutputs bool) []Slide {
	var slides []Slide

	for cellIdx, cell := range nb.Cells {
		cellNum := cellIdx + 1
		lines := notebook.SourceLines(cell.Source)

		// Collect output HTML
		var outputHTML strings.Builder
		if showOutputs && cell.CellType == "code" {
			for _, o := range cell.Outputs {
				outputHTML.WriteString(notebook.OutputHTML(o))
			}
		}
		outStr := outputHTML.String()

		// Split source into chunks of maxLines
		chunks := chunkLines(lines, maxLines)

		if len(chunks) <= 1 {
			// Single slide
			sid := fmt.Sprintf("%d", cellNum)
			src := ""
			if len(chunks) == 1 {
				src = strings.Join(chunks[0], "\n")
			}
			h := renderCell(sid, cell.CellType, src, outStr)
			slides = append(slides, Slide{ID: sid, CellIdx: cellNum, CellType: cell.CellType, HTML: h})
		} else {
			// Multiple slides: 5-a, 5-b, ...
			for i, chunk := range chunks {
				letter := string(rune('a' + i))
				sid := fmt.Sprintf("%d-%s", cellNum, letter)
				// Only put outputs on the last chunk
				chunkOut := ""
				if i == len(chunks)-1 {
					chunkOut = outStr
				}
				src := strings.Join(chunk, "\n")
				h := renderCell(sid, cell.CellType, src, chunkOut)
				slides = append(slides, Slide{ID: sid, CellIdx: cellNum, CellType: cell.CellType, HTML: h})
			}
		}
	}
	return slides
}

func chunkLines(lines []string, maxLines int) [][]string {
	if len(lines) == 0 {
		return [][]string{lines}
	}
	var chunks [][]string
	for i := 0; i < len(lines); i += maxLines {
		end := i + maxLines
		if end > len(lines) {
			end = len(lines)
		}
		chunks = append(chunks, lines[i:end])
	}
	return chunks
}

func renderCell(id, cellType, source, outputHTML string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`<div class="slide" id="slide-%s">`, id))
	b.WriteString(fmt.Sprintf(`<div class="slide-label">Cell %s</div>`, id))

	switch cellType {
	case "markdown":
		b.WriteString(`<div class="cell-md">`)
		b.WriteString(markdown.RenderMarkdown(source))
		b.WriteString(`</div>`)
	case "code":
		b.WriteString(`<div class="cell-code">`)
		b.WriteString(fmt.Sprintf(`<pre><code class="language-python">%s</code></pre>`, util.EscapeHTML(source)))
		b.WriteString(`</div>`)
	default:
		b.WriteString(`<div class="cell-raw">`)
		b.WriteString(fmt.Sprintf(`<pre>%s</pre>`, util.EscapeHTML(source)))
		b.WriteString(`</div>`)
	}

	if outputHTML != "" {
		b.WriteString(`<div class="cell-output">`)
		b.WriteString(`<div class="output-label">Output</div>`)
		b.WriteString(outputHTML)
		b.WriteString(`</div>`)
	}

	b.WriteString(`</div>`) // .slide
	return b.String()
}
