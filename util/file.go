package util

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type File struct {
	Path  string
	Name  string
	Hash  string
	Rerun bool
}

func (f *File) Read() (string, error) {
	b, err := os.ReadFile(f.Path)

	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (f *File) Head() (string, error) {
	content, err := f.Read()

	if err != nil {
		return "", err
	}

	return HeadContent(content), nil
}

func PathsToFiles(paths []string) ([]*File, error) {
	files := []*File{}

	for _, p := range paths {
		hash, err := Hash(p)

		if err != nil {
			return nil, fmt.Errorf("failed to calculate hash: %s: %w", p, err)
		}

		f := &File{
			Path: p,
			Name: filepath.Base(p),
			Hash: hash,
		}

		files = append(files, f)
	}

	slices.SortFunc(files, func(a, b *File) int { return strings.Compare(a.Name, b.Name) })

	return files, nil
}
