# splitbook

A minimalist command-line tool written in Go to extract specific page ranges from a PDF file.

## Usage

```sh
splitbook <pdf-file> <page-range>
```

### Examples

The tool accepts various formats for the page range:

```sh
splitbook document.pdf "118 - 122"
splitbook document.pdf 118 to 122
splitbook document.pdf 118-122
splitbook document.pdf 118
```

The output file is saved in the same directory as the input file, with the requested range appended to the filename (e.g., `document_118-122.pdf`).

## Building from source

To build the executable manually:

```sh
go build -o splitbook main.go
```

To build for all supported targets (Windows, macOS Intel/Apple Silicon, Linux):

```sh
./build.sh
```

The resulting binaries will be placed in the `bin/` directory.

## Dependencies

- [github.com/pdfcpu/pdfcpu](https://github.com/pdfcpu/pdfcpu)

### Recommended: Ghostscript

For the best performance, especially when extracting from very large PDF files, `splitbook` relies on [Ghostscript](https://ghostscript.com/releases/index.html).

If Ghostscript (`gs`, `gswin64c.exe`, etc.) is available on the system `PATH` or bundled in the exact same directory as the `splitbook` executable, the tool will:
- Extract pages significantly faster.
- Provide a real-time terminal progress bar (when extracting > 5 pages).

*Note: If Ghostscript is not found, `splitbook` will gracefully fall back to its internal `pdfcpu` engine.*
