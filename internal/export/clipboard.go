package export

import (
	"fmt"

	"github.com/atotto/clipboard"
)

// CopyToClipboard copies the provided content to the system clipboard
func CopyToClipboard(content string) error {
	if content == "" {
		return fmt.Errorf("cannot copy empty content to clipboard")
	}

	err := clipboard.WriteAll(content)
	if err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	return nil
}
