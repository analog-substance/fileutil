package fileutil

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	DefaultDirPerms  fs.FileMode = 0755
	DefaultFilePerms fs.FileMode = 0644
)

func HasStdin() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	mode := stat.Mode()

	isPipedFromChrDev := (mode & os.ModeCharDevice) == 0
	isPipedFromFIFO := (mode & os.ModeNamedPipe) != 0

	return isPipedFromChrDev || isPipedFromFIFO
}

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

func IsSameFile(p1 string, p2 string) bool {
	p1Info, err := os.Stat(p1)
	if err != nil {
		return false
	}

	p2Info, err := os.Stat(p2)
	if err != nil {
		return false
	}

	return os.SameFile(p1Info, p2Info)
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

func WriteString(path string, content string) error {
	return os.WriteFile(path, []byte(content), DefaultFilePerms)
}

func CopyFile(src string, dest string) error {
	if IsSameFile(src, dest) {
		return errors.New("source and destination are the same file")
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	srcInfo, _ := srcFile.Stat()

	if !srcInfo.Mode().IsRegular() {
		return fmt.Errorf("non-regular source file %s (%q)", srcInfo.Name(), srcInfo.Mode().String())
	}

	var destFile *os.File
	if FileExists(dest) { // dest is path/to/existing/file
		destFile, err = os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, srcInfo.Mode().Perm())
	} else if DirExists(dest) { // dest is path/to/existing/dir
		destFile, err = os.OpenFile(filepath.Join(dest, srcInfo.Name()), os.O_CREATE|os.O_WRONLY, srcInfo.Mode().Perm())
	} else if DirExists(filepath.Dir(dest)) { // dest is path/to/existing/dir/non_existent_file
		destFile, err = os.OpenFile(filepath.Join(filepath.Dir(dest), filepath.Base(dest)), os.O_CREATE|os.O_WRONLY, srcInfo.Mode().Perm())
	} else { // dest is to a path that doesn't exist
		err = errors.New("destination path doesn't exist")
	}

	if err != nil {
		return err
	}
	defer destFile.Close()

	written, err := io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	if written != srcInfo.Size() {
		return errors.New("error writing data to destination file")
	}

	return nil
}
