package init

import (
	"os"
)

func CanBeInitialized(directory string) bool {
	_, err := os.Stat(directory)
	if err != nil && os.IsNotExist(err) {
		return true
	}

	return false
}

func HasBeenInitialized(directory string) bool {
	return false
}

func Journal(directory string) error {
	return nil
}
