package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ConvertToBytes(size int, unit string) (int, error) {
	switch unit {
	case "B":
		return size, nil
	case "K":
		return size * 1024, nil
	case "M":
		return size * 1024 * 1024, nil
	default:
		return 0, errors.New("invalid unit")
	}
}

var alphabet = []string{
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
	"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
}

var LogedUserID int32 = 1
var LogedUserGroupID int32 = 1

var PathToLetter = make(map[string]string)

var nextLetterIndex = 0

var PathToPartitionCount = make(map[string]int)

func GetLetter(path string) (string, int, error) {
	if _, exists := PathToLetter[path]; !exists {
		if nextLetterIndex < len(alphabet) {
			PathToLetter[path] = alphabet[nextLetterIndex]
			PathToPartitionCount[path] = 0
			nextLetterIndex++
		} else {
			fmt.Println("Error: no hay más letras disponibles para asignar")
			return "", 0, errors.New("no hay más letras disponibles para asignar")
		}
	}

	PathToPartitionCount[path]++
	nextIndex := PathToPartitionCount[path]

	return PathToLetter[path], nextIndex, nil
}

func CreateParentDirs(path string) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error al crear las carpetas padre: %v", err)
	}
	return nil

}

func GetFileNames(path string) (string, string) {
	dir := filepath.Dir(path)
	baseName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	dotFileName := filepath.Join(dir, baseName+".dot")
	outpuImage := path
	return dotFileName, outpuImage
}

func GetParentDirectories(path string) ([]string, string) {
	path = filepath.Clean(path)
	components := strings.Split(path, string(filepath.Separator))
	var parentDirs []string
	for i := 1; i < len(components)-1; i++ {
		parentDirs = append(parentDirs, components[i])
	}
	destDir := components[len(components)-1]
	return parentDirs, destDir
}

func RemoveElement[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		return slice
	}
	return append(slice[:index], slice[index+1:]...)
}

func SplitStringIntoChunks(s string) []string {
	var chunks []string
	for i := 0; i < len(s); i += 64 {
		end := i + 64
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
}

func First[T any](slice []T) (T, error) {
	if len(slice) == 0 {
		var zero T
		return zero, errors.New("el slice está vacío")
	}
	return slice[0], nil
}

func GetNameByPath(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

var DiskLetters = []string{"A", "B", "C", "D"}
var UsedDiskLetters = make(map[string]bool)

func GetNextAvailableDiskLetter() (string, error) {
	for _, letter := range DiskLetters {
		if !UsedDiskLetters[letter] {
			UsedDiskLetters[letter] = true
			return letter, nil
		}
	}
	return "", errors.New("no hay letras de disco disponibles")
}

func ReleaseDiskLetter(letter string) {
	delete(UsedDiskLetters, letter)
}

func GetDiskLetterFromPath(path string) string {
	baseName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	for _, letter := range DiskLetters {
		if baseName == letter {
			return letter
		}
	}
	return ""
}
