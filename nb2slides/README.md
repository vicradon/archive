# nb2slides — Jupyter Notebook → PDF Slides

A CLI tool that converts `.ipynb` notebooks into PDF slides. Each cell becomes a slide. If a cell is too long it is automatically split into labelled sub-slides (e.g. **5-a**, **5-b**, ...).

---

## Requirements

- **Go 1.18+** — https://go.dev/dl/
- **Chromium or Google Chrome** (optional, for direct PDF export)  
  If not available the tool writes HTML files you print-to-PDF from a browser.

---

## Build

```bash
cd nb2slides
go build -o nb2slides .
```

On Windows:

```powershell
go build -o nb2slides.exe .
```

---

## Usage

### Direct PDF export (requires Chromium/Chrome)

```bash
./nb2slides -i notebook.ipynb -o output_dir
```

This writes one PDF per slide into `output_dir/`:

```
output_dir/
  slide_1.pdf
  slide_2.pdf
  slide_5_a.pdf   ← cell 5 was too long, split into a/b/c
  slide_5_b.pdf
  slide_5_c.pdf
  all_slides.html ← combined preview
```

### Browser preview (no Chromium required)

```bash
./nb2slides -i notebook.ipynb -serve
```

Opens `http://localhost:8080` in your browser. Use **Ctrl+P → Save as PDF** with:
- Destination: Save as PDF
- Layout: Landscape (or Portrait, your choice)
- Margins: None
- ☑ Background graphics

### HTML fallback (when Chromium is absent)

When you run without `-serve` and Chromium isn't found, the tool writes:

```
output_dir/
  index.html        ← start here
  all_slides.html   ← full deck, print this
  slide_1.html
  slide_2.html
  ...
```

Open `all_slides.html` and print.

---

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-i` | — | Input notebook path **(required)** |
| `-o` | `slides_output` | Output directory |
| `-max-lines` | `50` | Source lines per slide before splitting |
| `-outputs` | `true` | Include cell outputs |
| `-serve` | `false` | Run preview server |
| `-port` | `8080` | Port for `-serve` mode |

---

## Examples

```bash
# Standard export, split at 40 lines, hide outputs
./nb2slides -i analysis.ipynb -o analysis_slides -max-lines 40 -outputs=false

# Preview server on port 9000
./nb2slides -i analysis.ipynb -serve -port 9000
```

---

## Slide splitting rules

- Cell **5** with 120 lines and `-max-lines 50` → slides **5-a** (lines 1-50), **5-b** (51-100), **5-c** (101-120)
- Cell outputs only appear on the **last** sub-slide of a split cell
- Markdown cells are rendered with headers, lists, inline math, and code blocks

---

## Supported output types

| Type | Rendering |
|------|-----------|
| `stream` (stdout/stderr) | Monospace pre block, stderr shown in red |
| `execute_result` / `display_data` — text | Monospace pre block |
| `execute_result` / `display_data` — image/png or jpeg | Embedded inline image |
| `execute_result` / `display_data` — text/html | Rendered HTML (tables etc.) |
| `error` | Red pre block with ANSI codes stripped |
