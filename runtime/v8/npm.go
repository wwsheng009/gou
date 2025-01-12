package v8

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/yaoapp/gou/application"
)

type PackageInfo struct {
	Main   string `json:"main"`
	Module string `json:"module"`
}

func checkImportFilePath(file string) string {
	//maybe folder such like 'xx.js/index.js'
	suffixes := []string{".d.ts", ".ts", ".js", ".mjs"}
	if isDirectory(file) || !hasValidSuffix(file, suffixes) {
		if checkFileExist(filepath.Join(file, "package.json")) {
			if packageInfo, err := getPackageEntryFile(filepath.Join(file, "package.json")); err == nil {
				if updatedFile := resolvePackageEntry(file, packageInfo, suffixes); updatedFile != "" {
					return updatedFile
				}
			}
		}

		// Check for file with extensions
		if newFile, found := checkFileWithSuffixes(file, suffixes); found {
			return newFile
		}
		// Check for index files
		if newIndex, found := checkIndexFiles(file, suffixes); found {
			return newIndex
		}
	}
	return file
}

func resolvePackageEntry(file string, packageInfo PackageInfo, suffixes []string) string {
	entries := []string{packageInfo.Module, packageInfo.Main}

	for _, entry := range entries {
		if entry != "" {
			filePath := filepath.Join(file, entry)
			if !hasValidSuffix(filePath, suffixes) {
				if newFile, found := checkFileWithSuffixes(filePath, suffixes); found {
					return newFile
				}
				if newIndex, found := checkIndexFiles(filePath, suffixes); found {
					return newIndex
				}
			}
			if checkFileExist(filePath) {
				return filePath
			}
		}
	}
	return ""
}

func hasValidSuffix(file string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(file, suffix) {
			return true
		}
	}
	return false
}

func checkFileExist(filePath string) bool {
	if exist, _ := application.App.Exists(filePath); exist && !isDirectory(filePath) {
		return true
	}
	return false
}
func isDirectory(file string) bool {
	info, err := os.Stat(file)
	if err != nil {
		return false // If there's an error (e.g., file doesn't exist), return the error
	}

	// Check if the path is not a directory
	return info.IsDir()
}
func checkFileWithSuffixes(basePath string, suffixes []string) (string, bool) {
	for _, suffix := range suffixes {
		filePath := basePath + suffix
		if checkFileExist(filePath) {
			return filePath, true
		}
	}
	return "", false
}

func checkIndexFiles(path string, suffixes []string) (string, bool) {
	for _, suffix := range suffixes {
		indexPath := filepath.Join(path, "index"+suffix)
		if checkFileExist(indexPath) {
			return indexPath, true
		}
	}
	return "", false
}

// get the entry file for the package
func getPackageEntryFile(file string) (PackageInfo, error) {
	// Parse the JSON
	var pkg PackageInfo
	f, err := os.Open(file)
	if err != nil {
		return pkg, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Read the file contents
	bytes, err := io.ReadAll(f)
	if err != nil {
		return pkg, fmt.Errorf("failed to read file: %w", err)
	}

	err = json.Unmarshal(bytes, &pkg)
	if err != nil {
		return pkg, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return pkg, nil
}
