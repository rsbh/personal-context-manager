// Package search keeps a Bleve full-text index of contexts on disk.
// The index is a cache: md files are the source of truth.
package search

import "errors"

const defaultLimit = 20

var (
	ErrInvalidLimit = errors.New("invalid limit")
	ErrEmptyQuery   = errors.New("missing query")

	errNotImplemented = errors.New("not implemented")
)

// Doc is one context to index.
type Doc struct {
	ID          string
	Scope       string
	Project     string
	Section     string
	Name        string
	Title       string
	Description string
	Content     string
	Tags        []string
}

// Query selects hits. Empty Scope, Project and Section match all.
// Limit 0 means defaultLimit. A negative limit is ErrInvalidLimit.
type Query struct {
	Q       string
	Scope   string
	Project string
	Section string
	Limit   int
}

// Hit is one search result. Title is the context name if the context has no title.
type Hit struct {
	ID          string  `json:"id"`
	Scope       string  `json:"scope"`
	Project     string  `json:"project,omitempty"`
	Section     string  `json:"section"`
	Name        string  `json:"name"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Snippet     string  `json:"snippet"`
	Score       float64 `json:"score"`
}

type Index struct{}

// Open opens the index in dir. needRebuild is true if the index is new
// or its schema version changed.
func Open(dir string) (idx *Index, needRebuild bool, err error) {
	return nil, false, errNotImplemented
}

// Index adds or replaces d.
func (i *Index) Index(d Doc) error { return errNotImplemented }

// Delete removes the doc with id.
func (i *Index) Delete(id string) error { return errNotImplemented }

// Search returns hits for q, best first.
func (i *Index) Search(q Query) ([]Hit, error) { return nil, errNotImplemented }

// Rebuild replaces all docs with docs and returns the count.
func (i *Index) Rebuild(docs []Doc) (int, error) { return 0, errNotImplemented }

func (i *Index) Close() error { return errNotImplemented }

// setSchemaVersion writes the schema version. Tests use it to fake an old index.
func (i *Index) setSchemaVersion(v string) error { return errNotImplemented }
