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
