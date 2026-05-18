package frontmatter

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

var separator = []byte("---")

// Parse splits a markdown file into YAML frontmatter metadata and body.
// If no frontmatter is found, metadata is nil and body is the full content.
func Parse(content []byte) (metadata map[string]interface{}, body string, err error) {
	content = bytes.TrimLeft(content, "\n\r")
	if !bytes.HasPrefix(content, separator) {
		return nil, string(content), nil
	}

	// Find the closing ---
	rest := content[len(separator):]
	rest = bytes.TrimLeft(rest, " \t")
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	} else if len(rest) > 1 && rest[0] == '\r' && rest[1] == '\n' {
		rest = rest[2:]
	}

	// Check if closing --- is at the very start (empty frontmatter)
	var yamlData []byte
	var afterClose []byte
	if bytes.HasPrefix(rest, separator) {
		yamlData = nil
		afterClose = rest[len(separator):]
	} else {
		end := bytes.Index(rest, []byte("\n---"))
		if end == -1 {
			// Try \r\n--- for Windows-style line endings
			end = bytes.Index(rest, []byte("\r\n---"))
			if end == -1 {
				// No closing separator, treat entire content as body
				return nil, string(content), nil
			}
			yamlData = rest[:end]
			afterClose = rest[end+5:] // skip \r\n---
		} else {
			yamlData = rest[:end]
			afterClose = rest[end+4:] // skip \n---
		}
	}

	// Skip trailing newline after closing ---
	if len(afterClose) > 0 && afterClose[0] == '\n' {
		afterClose = afterClose[1:]
	} else if len(afterClose) > 1 && afterClose[0] == '\r' && afterClose[1] == '\n' {
		afterClose = afterClose[2:]
	}

	metadata = make(map[string]interface{})
	if len(bytes.TrimSpace(yamlData)) > 0 {
		if err := yaml.Unmarshal(yamlData, &metadata); err != nil {
			return nil, "", fmt.Errorf("parsing frontmatter YAML: %w", err)
		}
	}

	return metadata, string(afterClose), nil
}

// Marshal reconstructs a markdown file from metadata and body.
// If metadata is nil or empty, returns just the body.
func Marshal(metadata map[string]interface{}, body string) ([]byte, error) {
	if len(metadata) == 0 {
		return []byte(body), nil
	}

	yamlData, err := yaml.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("marshaling frontmatter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(yamlData)
	buf.WriteString("---\n")
	if body != "" {
		buf.WriteString(body)
	}
	return buf.Bytes(), nil
}
