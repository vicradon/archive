package notebook

import (
	"encoding/json"
	"fmt"
	"nb2slides/internal/util"
	"os"
	"strings"
)

// Notebook top-level structure (supports nbformat 4)
type Notebook struct {
	NBFormat int    `json:"nbformat"`
	Cells    []Cell `json:"cells"`
}

type Cell struct {
	CellType string          `json:"cell_type"` // "markdown", "code", "raw"
	Source   json.RawMessage `json:"source"`    // can be string or []string
	Outputs  []Output        `json:"outputs"`
}

type Output struct {
	OutputType string                     `json:"output_type"` // "stream","display_data","execute_result","error"
	Name       string                     `json:"name"`        // stdout/stderr for stream
	Text       json.RawMessage            `json:"text"`
	Data       map[string]json.RawMessage `json:"data"`
	Traceback  []string                   `json:"traceback"`
	EValue     string                     `json:"evalue"`
	EName      string                     `json:"ename"`
}

func ParseNotebook(path string) (*Notebook, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	var nb Notebook
	if err := json.Unmarshal(data, &nb); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}
	return &nb, nil
}

// SourceLines decodes the "source" field which can be a plain string OR []string
func SourceLines(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	// Try []string first
	var arr []string
	if err := json.Unmarshal(raw, &arr); err == nil {
		// Each element may contain embedded newlines
		var lines []string
		for _, s := range arr {
			lines = append(lines, strings.Split(s, "\n")...)
		}
		return lines
	}
	// Try plain string
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.Split(s, "\n")
	}
	return nil
}

// OutputText extracts printable text from an Output
func OutputText(o Output) string {
	switch o.OutputType {
	case "stream":
		lines := SourceLines(o.Text)
		return strings.Join(lines, "\n")
	case "execute_result", "display_data":
		// prefer text/plain
		if raw, ok := o.Data["text/plain"]; ok {
			lines := SourceLines(raw)
			return strings.Join(lines, "\n")
		}
		// fallback text/html stripped
		if raw, ok := o.Data["text/html"]; ok {
			lines := SourceLines(raw)
			return util.StripHTML(strings.Join(lines, "\n"))
		}
	case "error":
		// strip ANSI escape codes from traceback
		tb := strings.Join(o.Traceback, "\n")
		return util.StripANSI(tb)
	}
	return ""
}

// OutputHTML returns richer HTML for an output (may include image tags)
func OutputHTML(o Output) string {
	switch o.OutputType {
	case "stream":
		class := "out-stdout"
		if o.Name == "stderr" {
			class = "out-stderr"
		}
		lines := SourceLines(o.Text)
		return fmt.Sprintf(`<pre class="%s">%s</pre>`, class, util.EscapeHTML(strings.Join(lines, "\n")))
	case "execute_result", "display_data":
		// image/png
		if raw, ok := o.Data["image/png"]; ok {
			var b64 string
			_ = json.Unmarshal(raw, &b64)
			b64 = strings.ReplaceAll(b64, "\n", "")
			return fmt.Sprintf(`<div class="out-image"><img src="data:image/png;base64,%s"/></div>`, b64)
		}
		// image/jpeg
		if raw, ok := o.Data["image/jpeg"]; ok {
			var b64 string
			_ = json.Unmarshal(raw, &b64)
			b64 = strings.ReplaceAll(b64, "\n", "")
			return fmt.Sprintf(`<div class="out-image"><img src="data:image/jpeg;base64,%s"/></div>`, b64)
		}
		// text/html
		if raw, ok := o.Data["text/html"]; ok {
			lines := SourceLines(raw)
			return fmt.Sprintf(`<div class="out-html">%s</div>`, strings.Join(lines, "\n"))
		}
		// text/plain
		if raw, ok := o.Data["text/plain"]; ok {
			lines := SourceLines(raw)
			return fmt.Sprintf(`<pre class="out-stdout">%s</pre>`, util.EscapeHTML(strings.Join(lines, "\n")))
		}
	case "error":
		tb := strings.Join(o.Traceback, "\n")
		return fmt.Sprintf(`<pre class="out-stderr">%s</pre>`, util.EscapeHTML(util.StripANSI(tb)))
	}
	return ""
}
