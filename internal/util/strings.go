package util

import "strings"

// RemoveExtension removes the file extension from a filename
func RemoveExtension(filename string) string {
	idx := strings.LastIndex(filename, ".")
	if idx > 0 {
		return filename[:idx]
	}
	return filename
}

// CleanTitle cleans a title by replacing dots and underscores with spaces and trimming
func CleanTitle(title string) string {
	title = strings.ReplaceAll(title, ".", " ")
	title = strings.ReplaceAll(title, "_", " ")
	return strings.TrimSpace(title)
}

// ContainsExtension checks if ext is in the provided extensions slice (case-insensitive)
func ContainsExtension(extensions []string, ext string) bool {
	ext = strings.ToLower(ext)
	for _, e := range extensions {
		if e == ext {
			return true
		}
	}
	return false
}
