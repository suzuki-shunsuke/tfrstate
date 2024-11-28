package run

import (
	"encoding/json"
	"fmt"
	"io"
)

func output(changes []*Change, stdout io.Writer) error {
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(changes); err != nil {
		return fmt.Errorf("encode the result as JSON: %w", err)
	}
	return nil
}
