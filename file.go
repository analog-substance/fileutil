package fileutil

import (
	"bufio"
	"io"
	"io/fs"
	"os"
)

const (
	DefaultDirPerms  fs.FileMode = 0755
	DefaultFilePerms fs.FileMode = 0644
)

func ReadLines(path string) ([]string, error) {
	var lines []string

	c, err := ReadLineByLine(path)
	if err != nil {
		return lines, err
	}

	for line := range c {
		lines = append(lines, line)
	}

	return lines, nil
}

func ReadFileLines(r io.Reader) []string {
	var lines []string

	for line := range ReadFileLineByLine(r) {
		lines = append(lines, line)
	}

	return lines
}

func ReadLineByLine(path string) (chan string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	// Not sure if this is the best way to re-use ReadLineByLine
	c := make(chan string)
	go func() {
		defer file.Close()

		for s := range ReadFileLineByLine(file) {
			c <- s
		}
		close(c)
	}()

	return c, nil
}

func ReadFileLineByLine(r io.Reader) chan string {
	c := make(chan string)
	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			c <- scanner.Text()
		}
		close(c)
	}()

	return c
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func DirExists(dir string) bool {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func MkdirAll(dirs ...string) []error {
	var errors []error
	for _, dir := range dirs {
		err := os.MkdirAll(dir, DefaultDirPerms)
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func WriteFileLines(w io.Writer, lines []string) error {
	writer := bufio.NewWriter(w)
	for _, data := range lines {
		_, err := writer.WriteString(data + "\n")
		if err != nil {
			return err
		}
	}

	writer.Flush()
	return nil
}

func WriteLines(path string, lines []string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, DefaultFilePerms)
	if err != nil {
		return err
	}
	defer file.Close()

	return WriteFileLines(file, lines)
}
