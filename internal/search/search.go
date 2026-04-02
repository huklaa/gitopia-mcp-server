package search

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Match represents a single search match.
type Match struct {
	FilePath    string `json:"file_path"`
	LineNumber  uint64 `json:"line_number"`
	LineContent string `json:"line_content"`
}

// RipgrepMatch represents the structure of a JSON line from ripgrep's output.
type RipgrepMatch struct {
	Type string `json:"type"`
	Data struct {
		Path       RipgrepText `json:"path"`
		Lines      RipgrepText `json:"lines"`
		LineNumber uint64      `json:"line_number"`
	} `json:"data"`
}

type RipgrepText struct {
	Text string `json:"text"`
}

// CodeSearch performs a search using ripgrep for a given query in a repository.
func CodeSearch(repoPath, query string, caseSensitive bool) ([]Match, error) {
	args := []string{
		"--json",
		query,
		".", // Search the current directory
	}
	if !caseSensitive {
		args = append([]string{"-i"}, args...)
	}

	cmd := exec.Command("rg", args...)
	cmd.Dir = repoPath // Set the working directory for the command

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error creating stdout pipe for ripgrep: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("error starting ripgrep: %w. Is ripgrep installed on the server?", err)
	}

	var matches []Match
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		var rgMatch RipgrepMatch
		if err := json.Unmarshal([]byte(line), &rgMatch); err != nil {
			// Ignore lines that aren't valid JSON, like context separators.
			continue
		}

		if rgMatch.Type == "match" {
			matches = append(matches, Match{
				FilePath:    rgMatch.Data.Path.Text,
				LineNumber:  rgMatch.Data.LineNumber,
				LineContent: strings.TrimSpace(rgMatch.Data.Lines.Text),
			})
		}
	}

	if err := cmd.Wait(); err != nil {
		// ripgrep exits with status 1 if no matches are found, which is not a true error for us.
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return matches, nil // No matches found is a valid result.
		}
		return nil, fmt.Errorf("error waiting for ripgrep: %w", err)
	}

	return matches, nil
}
