package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var scanDir string

// Paths to scan in the Rails application
var scanPaths = []string{
	"cmd/api",
	"internal",
	"db",
	"templates",
	"permissions",
}

// DirectoryScanner handles scanning of Rails directories
type DirectoryScanner struct {
	rootPath string
}

// NewDirectoryScanner creates a new scanner instance
func NewDirectoryScanner(rootPath string) *DirectoryScanner {
	return &DirectoryScanner{
		rootPath: rootPath,
	}
}

// ScanDirectories processes all configured paths
func (ds *DirectoryScanner) ScanDirectories() error {
	if scanDir != "." {
		fullPath := filepath.Join(ds.rootPath)
		if _, err := os.Stat(fullPath); err == nil {
			if err := ds.scanDirectory(fullPath); err != nil {
				return fmt.Errorf("error scanning directory %s: %w", fullPath, err)
			}
		}
		return nil
	}

	for _, path := range scanPaths {
		fullPath := filepath.Join(ds.rootPath, path)

		if _, err := os.Stat(fullPath); err == nil {
			if err := ds.scanDirectory(fullPath); err != nil {
				return fmt.Errorf("error scanning directory %s: %w", fullPath, err)
			}
		}
	}
	return nil
}

// scanDirectory processes a single directory and its subdirectories
func (ds *DirectoryScanner) scanDirectory(path string) error {
	var files []string

	// Finds all .rb files in the directory and subdirectories
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && (strings.Contains(info.Name(), "node_modules") || strings.Contains(info.Name(), ".next")) {
			return filepath.SkipDir
		}

		if !info.IsDir() && (strings.Contains(info.Name(), "package-lock.json") || strings.Contains(info.Name(), "yarn.lock")) {
			return nil
		}

		if !info.IsDir() && (strings.HasSuffix(filePath, ".go") || strings.HasSuffix(filePath, ".json") || strings.HasSuffix(filePath, ".tsx") || strings.HasSuffix(filePath, ".ts") || strings.HasSuffix(filePath, ".jsx") || strings.HasSuffix(filePath, ".js") || strings.HasSuffix(filePath, ".yml") || strings.HasSuffix(filePath, ".perm") || strings.HasSuffix(filePath, ".tmpl")) {
			files = append(files, filePath)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking directory %s: %w", path, err)
	}

	// Sorts files for consistent output
	sort.Strings(files)

	// Processes each file
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("<!-- Error reading %s: %s -->\n", file, err)
			continue
		}

		relPath, err := filepath.Rel(ds.rootPath, file)
		if err != nil {
			fmt.Printf("<!-- Error getting relative path for %s: %s -->\n", file, err)
			continue
		}

		fmt.Printf("<file filename=\"%s\">\n", relPath)
		fmt.Printf("%s\n", strings.TrimSpace(string(content)))
		fmt.Printf("</file>\n\n")
	}

	return nil
}

func init() {
	flag.StringVar(&scanDir, "dir", ".", "Directory to scan")
	flag.Parse()
}

func main() {
	// Gets the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %s\n", err)
		os.Exit(1)
	}

	if scanDir != "." {
		currentDir = scanDir
	}

	scanner := NewDirectoryScanner(currentDir)

	fmt.Println("<!-- Generated Go Directory Structure -->")
	if err := scanner.ScanDirectories(); err != nil {
		fmt.Printf("Error scanning directories: %s\n", err)
		os.Exit(1)
	}
	fmt.Println("<!-- End of Go Directory Structure -->")
}
