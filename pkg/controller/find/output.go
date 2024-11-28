package find

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

func output(changes []*Change, stdout io.Writer, format string) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(changes); err != nil {
			return fmt.Errorf("encode the result as JSON: %w", err)
		}
		return nil
	case "markdown":
		lines := []string{
			"dir | file | outputs",
			"--- | --- | ---",
		}
		for _, change := range changes {
			for _, file := range change.Files {
				lines = append(lines, fmt.Sprintf("%s | %s | %s", change.Dir, file.Path, strings.Join(file.Outputs, ", ")))
			}
		}
		fmt.Fprintln(stdout, strings.Join(lines, "\n"))
		return nil
	}
	return errors.New("unsupported format")
}
