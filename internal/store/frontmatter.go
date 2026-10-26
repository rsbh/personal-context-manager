package store

import (
	"bytes"
	"fmt"

	"go.yaml.in/yaml/v3"
)

// Frontmatter is the YAML header of a context md file.
type Frontmatter struct {
	Title       string   `yaml:"title,omitempty" json:"title,omitempty"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Tags        []string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

func (f Frontmatter) isZero() bool {
	return f.Title == "" && f.Description == "" && len(f.Tags) == 0
}

var delim = []byte("---\n")

// parseFile splits a md file into frontmatter and body.
func parseFile(data []byte) (Frontmatter, string, error) {
	var fm Frontmatter
	if !bytes.HasPrefix(data, delim) {
		return fm, string(data), nil
	}
	rest := data[len(delim):]
	end := bytes.Index(rest, []byte("\n---\n"))
	if end < 0 {
		if !bytes.HasSuffix(rest, []byte("\n---")) {
			return fm, string(data), nil
		}
		end = len(rest) - len("\n---")
	}
	if err := yaml.Unmarshal(rest[:end], &fm); err != nil {
		return fm, "", fmt.Errorf("parse frontmatter: %w", err)
	}
	body := rest[end+len("\n---"):]
	body = bytes.TrimPrefix(body, []byte("\n"))
	return fm, string(body), nil
}

// Render joins frontmatter and body into md file content.
func Render(fm Frontmatter, body string) ([]byte, error) {
	if fm.isZero() {
		return []byte(body), nil
	}
	var b bytes.Buffer
	b.Write(delim)
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(fm); err != nil {
		return nil, fmt.Errorf("render frontmatter: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("render frontmatter: %w", err)
	}
	b.Write(delim)
	b.WriteString(body)
	return b.Bytes(), nil
}
