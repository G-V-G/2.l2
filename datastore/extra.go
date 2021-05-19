package datastore

import (
	"bufio"
	"os"
	"sort"
)

func searchValue(file *os.File, position int64) (string, error) {
		if _, err := file.Seek(position, 0); err != nil {
			return "",  err
		}
		reader := bufio.NewReader(file)
		return readValue(reader)
}

func getSortedKeys(index indexes) []string {
	keys := make([]string, 0)
	for key := range index {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}