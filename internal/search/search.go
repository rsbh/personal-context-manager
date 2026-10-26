// Package search keeps a Bleve full-text index of contexts on disk.
// The index is a cache: md files are the source of truth.
package search

import (
	"errors"
	"fmt"
	"html"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/standard"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/registry"
	"github.com/blevesearch/bleve/v2/search/highlight"
	htmlformat "github.com/blevesearch/bleve/v2/search/highlight/format/html"
	simplefrag "github.com/blevesearch/bleve/v2/search/highlight/fragmenter/simple"
	simplehl "github.com/blevesearch/bleve/v2/search/highlight/highlighter/simple"
	"github.com/blevesearch/bleve/v2/search/query"
	indexapi "github.com/blevesearch/bleve_index_api"
)

const (
	defaultLimit = 20
	snippetSize  = 150

	// schemaVersion changes when the mapping changes. An index with another
	// version is deleted and rebuilt.
	schemaVersion = "1"
	highlighter   = "pcm"
)

// Errors for invalid search input.
var (
	ErrInvalidLimit = errors.New("invalid limit")
	ErrEmptyQuery   = errors.New("missing query")

	schemaKey = []byte("schema_version")
)

// Field boosts: a title hit ranks above a description or tag hit,
// and these rank above a content hit.
var boosts = map[string]float64{
	"title":       3,
	"description": 2,
	"tags":        2,
	"content":     1,
}

func init() {
	err := registry.RegisterHighlighter(highlighter, func(map[string]any, *registry.Cache) (highlight.Highlighter, error) {
		return simplehl.NewHighlighter(
			simplefrag.NewFragmenter(snippetSize),
			htmlformat.NewFragmentFormatter("<mark>", "</mark>"),
			simplehl.DefaultSeparator), nil
	})
	if err != nil {
		panic(err)
	}
}

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
// Global with Project matches global contexts and that project.
// Limit 0 means defaultLimit. A negative limit is ErrInvalidLimit.
type Query struct {
	Q       string
	Scope   string
	Project string
	Global  bool
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

// Index is safe for concurrent use. Search runs during Rebuild on the old index.
type Index struct {
	dir string

	mu  sync.RWMutex // guards idx
	idx bleve.Index

	// writeMu orders writes. Rebuild holds it, so a write during a rebuild
	// waits and goes to the new index.
	writeMu sync.Mutex
}

// Open opens the index in dir. needRebuild is true if the index is new
// or its schema version changed.
func Open(dir string) (idx *Index, needRebuild bool, err error) {
	b, err := bleve.Open(dir)
	switch {
	case err == nil:
		v, verr := b.GetInternal(schemaKey)
		if verr == nil && string(v) == schemaVersion {
			return &Index{dir: dir, idx: b}, false, nil
		}
		b.Close()
	case !errors.Is(err, bleve.ErrorIndexPathDoesNotExist):
		// A broken index is a cache: drop it.
	}
	if err := os.RemoveAll(dir); err != nil {
		return nil, false, fmt.Errorf("remove index: %w", err)
	}
	b, err = create(dir)
	if err != nil {
		return nil, false, err
	}
	return &Index{dir: dir, idx: b}, true, nil
}

func create(dir string) (bleve.Index, error) {
	b, err := bleve.New(dir, newMapping())
	if err != nil {
		return nil, fmt.Errorf("create index: %w", err)
	}
	if err := b.SetInternal(schemaKey, []byte(schemaVersion)); err != nil {
		b.Close()
		return nil, fmt.Errorf("set schema version: %w", err)
	}
	return b, nil
}

func newMapping() mapping.IndexMapping {
	kw := mapping.NewKeywordFieldMapping()
	kw.Analyzer = keyword.Name
	kw.Store = true

	text := mapping.NewTextFieldMapping()
	text.Analyzer = standard.Name
	text.Store = true

	doc := mapping.NewDocumentStaticMapping()
	for _, f := range []string{"scope", "project", "section", "name"} {
		doc.AddFieldMappingsAt(f, kw)
	}
	for f := range boosts {
		doc.AddFieldMappingsAt(f, text)
	}

	m := bleve.NewIndexMapping()
	m.DefaultMapping = doc
	m.DefaultAnalyzer = standard.Name
	m.ScoringModel = indexapi.BM25Scoring
	return m
}

// record is the indexed form of a Doc.
type record struct {
	Scope       string   `json:"scope"`
	Project     string   `json:"project"`
	Section     string   `json:"section"`
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Tags        []string `json:"tags"`
}

func toRecord(d Doc) record {
	return record{d.Scope, d.Project, d.Section, d.Name, d.Title, d.Description, d.Content, d.Tags}
}

// Index adds or replaces d.
func (i *Index) Index(d Doc) error {
	i.writeMu.Lock()
	defer i.writeMu.Unlock()
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.idx.Index(d.ID, toRecord(d))
}

// Delete removes the doc with id.
func (i *Index) Delete(id string) error {
	i.writeMu.Lock()
	defer i.writeMu.Unlock()
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.idx.Delete(id)
}

// Rebuild replaces all docs with docs and returns the count.
// It builds a new index next to the old one and then swaps them.
func (i *Index) Rebuild(docs []Doc) (int, error) {
	i.writeMu.Lock()
	defer i.writeMu.Unlock()

	tmp := i.dir + ".tmp"
	if err := os.RemoveAll(tmp); err != nil {
		return 0, fmt.Errorf("remove %s: %w", tmp, err)
	}
	b, err := create(tmp)
	if err != nil {
		return 0, err
	}
	batch := b.NewBatch()
	for _, d := range docs {
		if err := batch.Index(d.ID, toRecord(d)); err != nil {
			b.Close()
			return 0, err
		}
	}
	if err := b.Batch(batch); err != nil {
		b.Close()
		return 0, err
	}
	if err := b.Close(); err != nil {
		return 0, err
	}

	i.mu.Lock()
	defer i.mu.Unlock()
	if err := i.idx.Close(); err != nil {
		return 0, err
	}
	if err := os.RemoveAll(i.dir); err != nil {
		return 0, fmt.Errorf("remove index: %w", err)
	}
	if err := os.Rename(tmp, i.dir); err != nil {
		return 0, fmt.Errorf("swap index: %w", err)
	}
	if i.idx, err = bleve.Open(i.dir); err != nil {
		return 0, fmt.Errorf("open index: %w", err)
	}
	return len(docs), nil
}

// Search returns hits for q, best first.
func (i *Index) Search(q Query) ([]Hit, error) {
	if q.Limit < 0 {
		return nil, ErrInvalidLimit
	}
	if q.Limit == 0 {
		q.Limit = defaultLimit
	}
	bq, err := buildQuery(q)
	if err != nil {
		return nil, err
	}

	req := bleve.NewSearchRequestOptions(bq, q.Limit, 0, false)
	req.Fields = []string{"scope", "project", "section", "name", "title", "description", "content"}
	req.Highlight = bleve.NewHighlightWithStyle(highlighter)
	req.Highlight.AddField("content")

	i.mu.RLock()
	res, err := i.idx.Search(req)
	i.mu.RUnlock()
	if err != nil {
		return nil, err
	}

	hits := make([]Hit, 0, len(res.Hits))
	for _, m := range res.Hits {
		h := Hit{
			ID:          m.ID,
			Scope:       field(m.Fields, "scope"),
			Project:     field(m.Fields, "project"),
			Section:     field(m.Fields, "section"),
			Name:        field(m.Fields, "name"),
			Title:       field(m.Fields, "title"),
			Description: field(m.Fields, "description"),
			Score:       m.Score,
		}
		if h.Title == "" {
			h.Title = h.Name
		}
		if frags := m.Fragments["content"]; len(frags) > 0 {
			h.Snippet = frags[0]
		} else {
			// Same form as a fragment: HTML escaped.
			h.Snippet = html.EscapeString(cut(field(m.Fields, "content"), snippetSize))
		}
		hits = append(hits, h)
	}
	return hits, nil
}

func field(fields map[string]any, name string) string {
	s, _ := fields[name].(string)
	return s
}

// cut returns the first n runes of s.
func cut(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}

// Close closes the index.
func (i *Index) Close() error {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.idx.Close()
}

// setSchemaVersion writes the schema version. Tests use it to fake an old index.
func (i *Index) setSchemaVersion(v string) error {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.idx.SetInternal(schemaKey, []byte(v))
}

// buildQuery parses q.Q: plain words are optional (OR, more matches rank
// higher), +word must match, -word must not match, "a phrase" is exact.
// Plain words are fuzzy.
func buildQuery(q Query) (query.Query, error) {
	terms := parse(q.Q)
	if len(terms) == 0 {
		return nil, ErrEmptyQuery
	}
	bq := bleve.NewBooleanQuery()
	positive := false
	for _, t := range terms {
		tq := termQuery(t)
		switch t.op {
		case '+':
			bq.AddMust(tq)
			positive = true
		case '-':
			bq.AddMustNot(tq)
		default:
			bq.AddShould(tq)
			positive = true
		}
	}
	if !positive {
		// Only -terms: match all other docs.
		bq.AddMust(bleve.NewMatchAllQuery())
	}

	var filters []query.Query
	for f, v := range map[string]string{"scope": q.Scope, "section": q.Section} {
		if v != "" {
			filters = append(filters, fieldIs(f, v))
		}
	}
	if q.Project != "" {
		var pq query.Query = fieldIs("project", q.Project)
		if q.Global {
			pq = bleve.NewDisjunctionQuery(pq, fieldIs("scope", "global"))
		}
		filters = append(filters, pq)
	}
	if len(filters) > 0 {
		bq.AddFilter(bleve.NewConjunctionQuery(filters...))
	}
	return bq, nil
}

// fieldIs matches docs whose keyword field f is v.
func fieldIs(f, v string) *query.TermQuery {
	tq := bleve.NewTermQuery(v)
	tq.SetField(f)
	return tq
}

type term struct {
	op     byte // '+', '-' or 0
	text   string
	phrase bool
}

// parse splits s into terms. A quote starts a phrase that ends at the next
// quote or at the end of s.
func parse(s string) []term {
	var terms []term
	for {
		s = strings.TrimSpace(s)
		if s == "" {
			return terms
		}
		var t term
		if s[0] == '+' || s[0] == '-' {
			t.op, s = s[0], s[1:]
		}
		if strings.HasPrefix(s, `"`) {
			s = s[1:]
			end := strings.IndexByte(s, '"')
			if end < 0 {
				end = len(s)
			}
			t.text, t.phrase = s[:end], true
			s = s[min(end+1, len(s)):]
		} else {
			end := strings.IndexAny(s, " \t\n")
			if end < 0 {
				end = len(s)
			}
			t.text, s = s[:end], s[end:]
		}
		if strings.TrimSpace(t.text) != "" {
			terms = append(terms, t)
		}
	}
}

// termQuery matches t in all text fields, with field boosts.
func termQuery(t term) query.Query {
	fuzz := fuzziness(t.text)
	var qs []query.Query
	for f, boost := range boosts {
		if t.phrase {
			pq := bleve.NewMatchPhraseQuery(t.text)
			pq.SetField(f)
			pq.SetBoost(boost)
			qs = append(qs, pq)
			continue
		}
		mq := bleve.NewMatchQuery(t.text)
		mq.SetField(f)
		mq.SetBoost(boost)
		mq.SetFuzziness(fuzz)
		qs = append(qs, mq)
	}
	return bleve.NewDisjunctionQuery(qs...)
}

// fuzziness is the max edit distance: 1 for words of 5 runes or fewer, else 2.
func fuzziness(word string) int {
	if utf8.RuneCountInString(word) <= 5 {
		return 1
	}
	return 2
}
