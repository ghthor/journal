package case_0_static

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"code.google.com/p/go.tools/godoc/vfs/mapfs"
)

// This algorithm will only mkdir's of 1 depth
func mkDirsIn(directory string, paths []string) (dirMap map[string][]string, err error) {

	// collect directories in a map as the keys
	dirMap = make(map[string][]string, len(paths))

	for _, path := range paths {
		d := filepath.Dir(path)

		dirMap[d] = append(dirMap[d], filepath.Base(path))
	}

	// make the directories
	for d, _ := range dirMap {
		if d == "." {
			continue
		}

		err := os.Mkdir(filepath.Join(directory, d), 0777)
		if err != nil {
			return nil, err
		}
	}

	return dirMap, nil
}

func NewIn(directory string) (case_0_directory string, entries []string, err error) {
	// Build a slice of the file names from the static mapfs Files
	paths := make([]string, 0, len(Files))
	for path, _ := range Files {
		paths = append(paths, path)
	}

	// Build the directory tree from the files list
	dirMap, err := mkDirsIn(directory, paths)
	if err != nil {
		return "", nil, err
	}

	// Create and copy all the files into the directory
	vfs := mapfs.New(Files)

	for file, _ := range Files {
		fi, err := vfs.Stat(file)
		if err != nil {
			return "", nil, errors.New(fmt.Sprintf("error stating source file: %s %s", file, err))
		}

		if fi.IsDir() {
			continue
		}

		source, err := vfs.Open(file)
		if err != nil {
			return "", nil, errors.New(fmt.Sprintf("error opening source file: %s %s", file, err))
		}
		defer source.Close()

		dest, err := os.OpenFile(filepath.Join(directory, file), os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return "", nil, errors.New(fmt.Sprintf("error opening dest file: %s %s", file, err))
		}
		defer dest.Close()

		_, err = io.Copy(dest, source)
		if err != nil {
			return "", nil, err
		}
	}

	return filepath.Join(directory, "case_0"), dirMap["case_0"], nil
}
