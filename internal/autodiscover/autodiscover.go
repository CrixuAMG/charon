package autodiscover

import (
	"fmt"
	"os"
	"path/filepath"
)

func Scan(paths []string) ([]string, error) {
	var found []string

	for _, root := range paths {
		entries, err := os.ReadDir(root)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if entry.IsDir() {
				fmt.Println("DIR:", entry.Name())
				fullPath := filepath.Join(root, entry.Name())
				found = append(found, fullPath)
			}
		}
	}

	return found, nil
}

