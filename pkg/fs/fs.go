// Joseph Bursey

package fs

import (
	"bufio"
	"os"
	"strings"

	"fafo/pkg/log"
)

func Exists(pathname string) bool {
	if _, err := os.Stat(pathname); !os.IsNotExist(err) {
		return true
	}
	return false
}

func Mkdir(pathname string) error {
	if !Exists(pathname) {
		if err := os.MkdirAll(pathname, os.FileMode(0755)); err != nil {
			return err
		}
	}
	return nil
}

func Wc(filename string) (int, error) {
	count := 0
	file, err := os.Open(filename)
	if err != nil {
		return -1, err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		count++
	}
	file.Close()
	return count, nil
}

func chomp(str string) string {
	return strings.TrimSuffix(str, "\n")
}

func GetFileFromStdio(msg string) string {
	var filename string
	stdin := bufio.NewReader(os.Stdin)
	for {
		log.Logf(0, "%v: ", msg)
		filename, _ = stdin.ReadString('\n')
		filename = chomp(filename)
		if Exists(filename) {
			break
		} else {
			log.Logf(0, "File/directory \"%v\" does not exist! Try again?\n", filename)
		}
	}
	return filename
}
