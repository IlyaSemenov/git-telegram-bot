package telegram

import (
	"strings"
)

// FormatCommitMessage returns the first line of a commit message.
// If the message has multiple lines, it appends ellipsis to indicate truncation.
func FormatCommitMessage(message string) string {
	lines := strings.Split(strings.TrimSpace(message), "\n")
	firstLine := strings.TrimSpace(lines[0])

	if len(lines) > 1 {
		return firstLine + " (…)"
	}
	return firstLine
}
