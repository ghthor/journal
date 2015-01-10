package case_0_static

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	entryPkg "github.com/ghthor/journal/entry"
	"github.com/ghthor/journal/git"

	"code.google.com/p/go.tools/godoc/vfs/mapfs"
)

type entriesByDate []string

func (f entriesByDate) Len() int { return len(f) }
func (f entriesByDate) Less(i, j int) bool {
	iTime, err := time.Parse(entryPkg.FilenameLayout, f[i])
	if err != nil {
		panic(err)
	}

	jTime, err := time.Parse(entryPkg.FilenameLayout, f[j])
	if err != nil {
		panic(err)
	}

	return jTime.After(iTime)
}
func (f entriesByDate) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

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

// Copies all the static resources into directory
// with all the entries commited to the repository.
// The directory will also contain a reflog of
// the commits that fix will create to
// update the journal to the current format.
//
// The directory tree will be
/*
	directory/
	├── case_0
	│   ├── 2014-01-01-0000-EST
	│   ├── 2014-01-02-0000-EST
	│   ├── 2014-01-03-0000-EST
	│   ├── 2014-01-04-0000-EST
	│   ├── 2014-01-05-0000-EST
	│   ├── 2014-01-06-0000-EST
	│   └── 2014-01-07-0000-EST
	├── case_0.json
	└── case_0_fix_reflog
	    ├── 0
	    ├── 1
	    ├── 10
	    ├── 11
	    ├── 12
	    ├── 13
	    ├── 14
	    ├── 2
	    ├── 3
	    ├── 4
	    ├── 5
	    ├── 6
	    ├── 7
	    ├── 8
	    └── 9
*/
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

	// Initialize a git repository in the case_0/ directory and commit all the entries
	case_0_directory = filepath.Join(directory, "case_0")

	err = git.Init(case_0_directory)
	if err != nil {
		return "", nil, err
	}

	entries = dirMap["case_0"]
	sort.Sort(entriesByDate(entries))

	// Commit the entries into git
	for i, entryFilename := range entries {
		changes := git.NewChangesIn(case_0_directory)
		changes.Add(git.ChangedFile(entryFilename))
		changes.Msg = fmt.Sprintf("Commit Msg | Entry %d\n", i+1)
		err = changes.Commit()
		if err != nil {
			return "", nil, err
		}
	}

	return case_0_directory, entries, nil
}
