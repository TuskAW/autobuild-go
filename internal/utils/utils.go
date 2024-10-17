package utils

import (
	"os"
)

// EnsureDir ensures that a directory exists, if not, creates it
func EnsureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, os.ModePerm)
	}
	return nil
}
