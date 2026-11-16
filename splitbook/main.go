package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
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

	// Calculate start and end pages
	parts := strings.Split(rangeStr, "-")
	startPage, _ := strconv.Atoi(parts[0])
	endPage := startPage
	if len(parts) > 1 {
		endPage, _ = strconv.Atoi(parts[1])
	}

	// Swap if inverted
	if startPage > endPage {
		startPage, endPage = endPage, startPage
	}
	totalPages := endPage - startPage + 1

	// Normalize range string for filename
	if startPage == endPage {
		rangeStr = fmt.Sprintf("%d", startPage)
	} else {
		rangeStr = fmt.Sprintf("%d-%d", startPage, endPage)
	}

	ext := filepath.Ext(inFile)
	base := strings.TrimSuffix(filepath.Base(inFile), ext)
	dir := filepath.Dir(inFile)

	// Ensure consistent forward slashes in environments like WSL where a Windows binary
	// might receive a Unix-style path but try to join it with Windows backslashes.
	dir = filepath.ToSlash(dir)
	outFile := fmt.Sprintf("%s/%s_%s%s", dir, base, rangeStr, ext)
	
	// If the file was strictly in the current directory (no dir path), remove the leading slash
	if dir == "." {
		outFile = fmt.Sprintf("%s_%s%s", base, rangeStr, ext)
	}

	fmt.Printf("Extracting pages %s from %s to %s...\n", rangeStr, inFile, outFile)

	if extractWithGhostscript(inFile, outFile, startPage, endPage, totalPages) {
		fmt.Println("Extraction successful!")
		return
	}

	fmt.Println("Ghostscript progress extraction not available or failed. Falling back to pdfcpu...")

	// Fallback to pdfcpu with heavy validation disabled for large files
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed
	err = api.TrimFile(inFile, outFile, []string{rangeStr}, conf)
	if err != nil {
		fmt.Printf("Error extracting pages: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Extraction successful!")
}

func findGhostscript() string {
	// 1. Check same directory as the executable
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		// Check common names for bundled Ghostscript binaries
		names := []string{"gs", "gs.exe", "gswin64c.exe", "gs10060w64.exe", "gswin32c.exe"}
		for _, name := range names {
			gsPath := filepath.Join(exeDir, name)
			if _, err := os.Stat(gsPath); err == nil {
				return gsPath
			}
		}
	}

	// 2. Check system PATH using exec.LookPath
	names := []string{"gs", "gswin64c.exe", "gs10060w64.exe", "gswin32c.exe"}
	for _, name := range names {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}

	// 3. Fallback: If not found explicitly, let the OS try to resolve the most common names during exec.Command.
	// This happens sometimes on Windows where Go's LookPath doesn't see a recently updated PATH, but cmd.exe does.
	return "gswin64c.exe"
}

func extractWithGhostscript(inFile, outFile string, start, end, totalPages int) bool {
	gsPath := findGhostscript()
	if gsPath == "" {
		return false // gs not found
	}

	args := []string{
		"-dNOPAUSE",
		"-dBATCH",
		"-sDEVICE=pdfwrite",
		fmt.Sprintf("-dFirstPage=%d", start),
		fmt.Sprintf("-dLastPage=%d", end),
		fmt.Sprintf("-sOutputFile=%s", outFile),
		inFile,
	}

	cmd := exec.Command(gsPath, args...)

	// Ghostscript writes progress and messages to stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return false
	}

	if err := cmd.Start(); err != nil {
		return false
	}

	showProgress := totalPages > 5
	extractedCount := 0

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Page ") {
			extractedCount++
			if showProgress {
				percentage := float64(extractedCount) / float64(totalPages) * 100
				fmt.Printf("Progress: Extracted page %d of %d (%.0f%%)\r", extractedCount, totalPages, percentage)
			}
		}
	}

	if showProgress && extractedCount > 0 {
		fmt.Println() // Add a newline after the progress carriage return
	}

	err = cmd.Wait()
	if err != nil {
		os.Remove(outFile) // Clean up partial file on error
		return false
	}

	// Verify the file was actually created and has size
	info, err := os.Stat(outFile)
	if err != nil || info.Size() == 0 {
		return false
	}

	return true
}
