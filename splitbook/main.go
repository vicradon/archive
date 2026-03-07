package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: splitbook <pdf-file> <page-range>")
		fmt.Println("Example page ranges: '118-122', '118 - 122', '118 to 122', '118'")
		os.Exit(1)
	}

	inFile := os.Args[1]
	rangeStr := strings.Join(os.Args[2:], " ")

	rangeStr = strings.ToLower(rangeStr)
	rangeStr = strings.ReplaceAll(rangeStr, "to", "-")
	rangeStr = strings.ReplaceAll(rangeStr, " ", "")

	matched, err := regexp.MatchString(`^\d+(-\d+)?$`, rangeStr)
	if err != nil || !matched {
		fmt.Printf("Invalid page range format: %s\n", rangeStr)
		os.Exit(1)
	}

	ext := filepath.Ext(inFile)
	base := strings.TrimSuffix(filepath.Base(inFile), ext)
	dir := filepath.Dir(inFile)

	outFile := filepath.Join(dir, fmt.Sprintf("%s_%s%s", base, rangeStr, ext))

	fmt.Printf("Extracting pages %s from %s to %s...\n", rangeStr, inFile, outFile)

	err = api.TrimFile(inFile, outFile, []string{rangeStr}, nil)
	if err != nil {
		fmt.Printf("Error extracting pages: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Extraction successful!")
}
