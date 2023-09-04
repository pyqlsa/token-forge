// Package fileutil provides generic file utilities.
package fileutil

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// SpaceIndent is 4 spaces.
const SpaceIndent string = "    "

// ReadLines reads lines from a text file, ignoring blank lines and lines
// prefixed with "//" or "#" as comments.
func ReadLines(fileName string) ([]string, error) {
	file, err := os.Open(fileName) //#nosec:G304
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	var text []string
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		tmp := strings.TrimSpace(scanner.Text())
		tmp = stripComment(tmp, "//")
		tmp = stripComment(tmp, "#")
		if !(len(tmp) == 0) {
			text = append(text, tmp)
		}
	}

	if err := file.Close(); err != nil {
		return text, fmt.Errorf("failed to close file: %w", err)
	}

	return text, nil
}

// stripComment remove comments from the given line, denoted by the given
// separator; whole-line comments returned as empty string, and lines with
// in-line comments returned as the whitespace-trimmed content to the left of
// the separator.
func stripComment(line, sep string) string {
	tmp := line
	i := strings.Index(line, sep)
	if i >= 0 && i < len(line) {
		tmp = strings.TrimSpace(line[:i])
	}

	return tmp
}

// WriteJSON writes the interface as json to the given writer.
func WriteJSON(w io.Writer, v interface{}) error {
	json, err := json.MarshalIndent(v, "", SpaceIndent)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	fmt.Fprintf(w, "%s\n", string(json))

	return nil
}
