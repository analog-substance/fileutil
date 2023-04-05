package grep

import (
	"bufio"
	"io"
	"os"
	"regexp"

	"github.com/analog-substance/fileutil"
)

func Lines(path string, re *regexp.Regexp) ([]string, error) {
	var lines []string
	c, err := LineByLine(path, re)
	if err != nil {
		return nil, err
	}

	for line := range c {
		lines = append(lines, line)
	}

	return lines, nil
}

func FileLines(r io.Reader, re *regexp.Regexp) []string {
	var lines []string
	for line := range FileLineByLine(r, re) {
		lines = append(lines, line)
	}

	return lines
}

// LineByLine returns each line of the file matching the regex
func LineByLine(path string, re *regexp.Regexp) (chan string, error) {
	lineChan, err := fileutil.ReadLineByLine(path)
	if err != nil {
		return nil, err
	}

	matchChan := make(chan string)
	go func() {
		for line := range lineChan {
			if re.MatchString(line) {
				matchChan <- line
			}
		}
	}()

	return matchChan, nil
}

func FileLineByLine(r io.Reader, re *regexp.Regexp) chan string {
	lineChan := fileutil.ReadFileLineByLine(r)
	matchChan := make(chan string)
	go func() {
		for line := range lineChan {
			if re.MatchString(line) {
				matchChan <- line
			}
		}
	}()

	return matchChan
}

// Matches returns maximum n number of matches from the file.
//
//	n > 0: at most n matches
//	n == 0: the result is nil (zero matches)
//	n < 0: all matches
func Matches(path string, re *regexp.Regexp, n int) []string {
	if n == 0 {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	var matches []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if n > 0 && len(matches) == n {
			break
		}

		line := scanner.Text()
		if re.MatchString(line) {
			matches = append(matches, line)
		}
	}

	return matches
}

// Match returns true if any lines of the file match the regex
func Match(path string, re *regexp.Regexp) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	return FileMatch(file, re)
}

func FileMatch(r io.Reader, re *regexp.Regexp) bool {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if re.MatchString(scanner.Text()) {
			return true
		}
	}
	return false
}
