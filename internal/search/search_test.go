package search

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
)

var (
	goStyle = Doc{ID: "global/style/gostyle", Scope: "global", Section: "style", Name: "gostyle",
		Title: "Go", Description: "Go style", Content: "Use gofmt for formatting.", Tags: []string{"go"}}
	goReview = Doc{ID: "global/review/gorev", Scope: "global", Section: "review", Name: "gorev",
		Content: "Check gofmt in review."}
	apiGo = Doc{ID: "project/api/style/apigo", Scope: "project", Project: "api", Section: "style", Name: "apigo",
		Content: "Run gofmt in api."}
)

// newIndex opens a new index in a temp dir and adds docs.
func newIndex(t *testing.T, docs ...Doc) *Index {
	t.Helper()
	idx, _, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { idx.Close() })
	for _, d := range docs {
		if err := idx.Index(d); err != nil {
			t.Fatal(err)
		}
	}
	return idx
}

// ids runs q and returns the hit ids, best first.
func ids(t *testing.T, idx *Index, q Query) []string {
	t.Helper()
	hits, err := idx.Search(q)
	if err != nil {
		t.Fatalf("Search(%+v): %v", q, err)
	}
	out := make([]string, len(hits))
	for i, h := range hits {
		out[i] = h.ID
	}
	return out
}

func TestOpenNeedRebuild(t *testing.T) {
	dir := t.TempDir()
	idx, need, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !need {
		t.Error("new index: needRebuild = false, want true")
	}
	if err := idx.Close(); err != nil {
		t.Fatal(err)
	}

	idx, need, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	if need {
		t.Error("existing index: needRebuild = true, want false")
	}
}

func TestOpenSchemaChange(t *testing.T) {
	dir := t.TempDir()
	idx, _, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := idx.Index(goStyle); err != nil {
		t.Fatal(err)
	}
	if err := idx.setSchemaVersion("0"); err != nil {
		t.Fatal(err)
	}
	idx.Close()

	idx, need, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	if !need {
		t.Error("old schema: needRebuild = false, want true")
	}
	if got := ids(t, idx, Query{Q: "gofmt"}); len(got) != 0 {
		t.Errorf("old schema: hits = %v, want none", got)
	}
}

func TestPersist(t *testing.T) {
	dir := t.TempDir()
	idx, _, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := idx.Index(goStyle); err != nil {
		t.Fatal(err)
	}
	idx.Close()

	idx, _, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	if got := ids(t, idx, Query{Q: "gofmt"}); !slices.Equal(got, []string{goStyle.ID}) {
		t.Errorf("after reopen: hits = %v", got)
	}
}

func TestSearchFields(t *testing.T) {
	idx := newIndex(t, Doc{ID: "global/a/x", Scope: "global", Section: "a", Name: "x", Title: "Alpha"},
		Doc{ID: "global/a/y", Scope: "global", Section: "a", Name: "y", Description: "bravo"},
		Doc{ID: "global/a/z", Scope: "global", Section: "a", Name: "z", Tags: []string{"charlie"}},
		Doc{ID: "global/a/w", Scope: "global", Section: "a", Name: "w", Content: "delta"})
	for q, want := range map[string]string{"alpha": "global/a/x", "bravo": "global/a/y", "charlie": "global/a/z", "delta": "global/a/w"} {
		if got := ids(t, idx, Query{Q: q}); !slices.Equal(got, []string{want}) {
			t.Errorf("%q: hits = %v, want [%s]", q, got, want)
		}
	}
}

func TestHitFields(t *testing.T) {
	idx := newIndex(t, apiGo)
	hits, err := idx.Search(Query{Q: "gofmt"})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 {
		t.Fatalf("hits = %d, want 1", len(hits))
	}
	h := hits[0]
	if h.ID != apiGo.ID || h.Scope != "project" || h.Project != "api" || h.Section != "style" || h.Name != "apigo" {
		t.Errorf("hit = %+v", h)
	}
	if h.Title != "apigo" {
		t.Errorf("no title: Title = %q, want name %q", h.Title, "apigo")
	}
	if h.Score <= 0 {
		t.Errorf("Score = %v, want > 0", h.Score)
	}
	if !strings.Contains(h.Snippet, "<mark>gofmt</mark>") {
		t.Errorf("Snippet = %q, want <mark>gofmt</mark>", h.Snippet)
	}
}

func TestSnippetLength(t *testing.T) {
	long := strings.Repeat("word ", 200) + "needle " + strings.Repeat("word ", 200)
	idx := newIndex(t, Doc{ID: "global/a/x", Scope: "global", Section: "a", Name: "x", Content: long})
	hits, err := idx.Search(Query{Q: "needle"})
	if err != nil || len(hits) != 1 {
		t.Fatalf("hits = %v, err = %v", hits, err)
	}
	plain := strings.NewReplacer("<mark>", "", "</mark>", "").Replace(hits[0].Snippet)
	if n := len([]rune(plain)); n > 200 {
		t.Errorf("snippet length = %d, want about 150", n)
	}
}

func TestSnippetEscaped(t *testing.T) {
	idx := newIndex(t, Doc{ID: "global/a/x", Scope: "global", Section: "a", Name: "x", Title: "Tags",
		Content: "Use <b> & <i>."})
	for _, q := range []string{"tags", "use"} { // title only (no fragment), content (fragment)
		hits, err := idx.Search(Query{Q: q})
		if err != nil || len(hits) != 1 {
			t.Fatalf("%q: hits = %v, err = %v", q, hits, err)
		}
		if s := hits[0].Snippet; strings.Contains(s, "<b>") || !strings.Contains(s, "&lt;b&gt;") {
			t.Errorf("%q: Snippet = %q, want HTML escaped", q, s)
		}
	}
}

func TestFuzzy(t *testing.T) {
	idx := newIndex(t, Doc{ID: "global/a/x", Scope: "global", Section: "a", Name: "x", Content: "black formatting"})
	for _, tc := range []struct {
		q    string
		want bool
	}{
		{"blac", true},      // 4 chars, 1 edit
		{"bluuk", false},    // 5 chars, 2 edits
		{"formating", true}, // long, 1 edit
		{"frmating", true},  // long, 2 edits
		{"frmatin", false},  // long, 3 edits
	} {
		got := len(ids(t, idx, Query{Q: tc.q})) == 1
		if got != tc.want {
			t.Errorf("%q: hit = %v, want %v", tc.q, got, tc.want)
		}
	}
}

func TestOperators(t *testing.T) {
	idx := newIndex(t, goStyle, goReview)
	for _, tc := range []struct {
		q    string
		want []string
	}{
		{"+gofmt +review", []string{goReview.ID}},
		{"gofmt -review", []string{goStyle.ID}},
		{`"gofmt for formatting"`, []string{goStyle.ID}},
	} {
		if got := ids(t, idx, Query{Q: tc.q}); !slices.Equal(got, tc.want) {
			t.Errorf("%q: hits = %v, want %v", tc.q, got, tc.want)
		}
	}
}

func TestMultiWordOr(t *testing.T) {
	idx := newIndex(t, goStyle, goReview)
	got := ids(t, idx, Query{Q: "formatting review"})
	slices.Sort(got)
	if want := []string{goReview.ID, goStyle.ID}; !slices.Equal(got, want) {
		t.Errorf("hits = %v, want %v", got, want)
	}
}

func TestMoreWordsRankHigher(t *testing.T) {
	idx := newIndex(t, goStyle, goReview)
	got := ids(t, idx, Query{Q: "gofmt formatting"})
	if len(got) != 2 || got[0] != goStyle.ID {
		t.Errorf("hits = %v, want %s first", got, goStyle.ID)
	}
}

func TestTitleBoost(t *testing.T) {
	idx := newIndex(t,
		Doc{ID: "global/n/body-hit", Scope: "global", Section: "n", Name: "body-hit", Content: "Some testing tips."},
		Doc{ID: "global/n/title-hit", Scope: "global", Section: "n", Name: "title-hit", Title: "Testing", Content: "x"})
	got := ids(t, idx, Query{Q: "testing"})
	if len(got) != 2 || got[0] != "global/n/title-hit" {
		t.Errorf("hits = %v, want title-hit first", got)
	}
}

func TestFilters(t *testing.T) {
	idx := newIndex(t, goStyle, goReview, apiGo)
	for _, tc := range []struct {
		q    Query
		want []string
	}{
		{Query{Q: "gofmt", Section: "review"}, []string{goReview.ID}},
		{Query{Q: "gofmt", Project: "api"}, []string{apiGo.ID}},
		{Query{Q: "gofmt", Scope: "project"}, []string{apiGo.ID}},
		{Query{Q: "gofmt", Scope: "global", Section: "style"}, []string{goStyle.ID}},
	} {
		if got := ids(t, idx, tc.q); !slices.Equal(got, tc.want) {
			t.Errorf("%+v: hits = %v, want %v", tc.q, got, tc.want)
		}
	}
}

func TestProjectWithGlobal(t *testing.T) {
	webGo := Doc{ID: "projects/web/style/webgo", Scope: "project", Project: "web", Section: "style", Name: "webgo",
		Content: "Run gofmt in web."}
	idx := newIndex(t, goStyle, goReview, apiGo, webGo)
	for _, tc := range []struct {
		q    Query
		want []string
	}{
		{Query{Q: "gofmt", Project: "api", Global: true}, []string{goReview.ID, goStyle.ID, apiGo.ID}},
		{Query{Q: "gofmt", Project: "api", Global: true, Section: "style"}, []string{goStyle.ID, apiGo.ID}},
	} {
		got := ids(t, idx, tc.q)
		slices.Sort(got)
		if !slices.Equal(got, tc.want) {
			t.Errorf("%+v: hits = %v, want %v", tc.q, got, tc.want)
		}
	}
}

func TestLimit(t *testing.T) {
	var docs []Doc
	for i := range 25 {
		docs = append(docs, Doc{ID: fmt.Sprintf("global/a/d%d", i), Scope: "global", Section: "a",
			Name: fmt.Sprintf("d%d", i), Content: "gofmt"})
	}
	idx := newIndex(t, docs...)
	if got := ids(t, idx, Query{Q: "gofmt"}); len(got) != defaultLimit {
		t.Errorf("default: hits = %d, want %d", len(got), defaultLimit)
	}
	if got := ids(t, idx, Query{Q: "gofmt", Limit: 3}); len(got) != 3 {
		t.Errorf("limit 3: hits = %d", len(got))
	}
	if _, err := idx.Search(Query{Q: "gofmt", Limit: -1}); !errors.Is(err, ErrInvalidLimit) {
		t.Errorf("limit -1: err = %v, want ErrInvalidLimit", err)
	}
}

func TestEmptyQuery(t *testing.T) {
	idx := newIndex(t, goStyle)
	for _, q := range []string{"", "   "} {
		if _, err := idx.Search(Query{Q: q}); !errors.Is(err, ErrEmptyQuery) {
			t.Errorf("%q: err = %v, want ErrEmptyQuery", q, err)
		}
	}
}

func TestNoHit(t *testing.T) {
	idx := newIndex(t, goStyle)
	hits, err := idx.Search(Query{Q: "zzzqqq"})
	if err != nil {
		t.Fatal(err)
	}
	if hits == nil || len(hits) != 0 {
		t.Errorf("hits = %#v, want empty non-nil slice", hits)
	}
}

func TestIndexReplaces(t *testing.T) {
	idx := newIndex(t, goStyle)
	updated := goStyle
	updated.Content = "Use goimports."
	if err := idx.Index(updated); err != nil {
		t.Fatal(err)
	}
	if got := ids(t, idx, Query{Q: "gofmt"}); len(got) != 0 {
		t.Errorf("old word: hits = %v, want none", got)
	}
	if got := ids(t, idx, Query{Q: "goimports"}); !slices.Equal(got, []string{goStyle.ID}) {
		t.Errorf("new word: hits = %v", got)
	}
}

func TestDelete(t *testing.T) {
	idx := newIndex(t, goStyle)
	if err := idx.Delete(goStyle.ID); err != nil {
		t.Fatal(err)
	}
	if got := ids(t, idx, Query{Q: "gofmt"}); len(got) != 0 {
		t.Errorf("hits = %v, want none", got)
	}
}

func TestRebuild(t *testing.T) {
	dir := t.TempDir()
	idx, _, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := idx.Index(goStyle); err != nil {
		t.Fatal(err)
	}
	n, err := idx.Rebuild([]Doc{goReview, apiGo})
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("Rebuild = %d, want 2", n)
	}
	got := ids(t, idx, Query{Q: "gofmt"})
	slices.Sort(got)
	if want := []string{goReview.ID, apiGo.ID}; !slices.Equal(got, want) {
		t.Errorf("hits = %v, want %v", got, want)
	}
	idx.Close()

	// The rebuilt index is the one on disk.
	idx, need, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	if need {
		t.Error("after rebuild: needRebuild = true")
	}
	if got := ids(t, idx, Query{Q: "gofmt"}); len(got) != 2 {
		t.Errorf("after reopen: hits = %v", got)
	}
}

func TestOpenSchemaV1Rebuilds(t *testing.T) {
	dir := t.TempDir()
	idx, _, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := idx.setSchemaVersion("1"); err != nil {
		t.Fatal(err)
	}
	idx.Close()

	idx, need, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	if !need {
		t.Error("schema 1: needRebuild = false, want true")
	}
}

func TestNameSectionProjectMatch(t *testing.T) {
	idx := newIndex(t,
		Doc{ID: "global/notes/error-handling", Scope: "global", Section: "notes", Name: "error-handling", Content: "x"},
		Doc{ID: "global/review/pr", Scope: "global", Section: "review", Name: "pr", Content: "x"},
		Doc{ID: "project/billing/style/ctx", Scope: "project", Project: "billing", Section: "style", Name: "ctx", Content: "x"})
	for q, want := range map[string]string{
		"handling": "global/notes/error-handling",
		"review":   "global/review/pr",
		"billing":  "project/billing/style/ctx",
	} {
		got := ids(t, idx, Query{Q: q})
		if len(got) != 1 || got[0] != want {
			t.Errorf("Search(%q) = %v, want [%s]", q, got, want)
		}
	}
}

func TestNameBoost(t *testing.T) {
	idx := newIndex(t,
		Doc{ID: "global/n/other", Scope: "global", Section: "n", Name: "other", Content: "Some caching tips."},
		Doc{ID: "global/n/caching", Scope: "global", Section: "n", Name: "caching", Content: "x"})
	got := ids(t, idx, Query{Q: "caching"})
	if len(got) != 2 || got[0] != "global/n/caching" {
		t.Errorf("hits = %v, want caching first", got)
	}
}

func TestFilterStillExact(t *testing.T) {
	idx := newIndex(t,
		Doc{ID: "global/review/a", Scope: "global", Section: "review", Name: "a", Content: "gofmt"},
		Doc{ID: "global/code-review/b", Scope: "global", Section: "code-review", Name: "b", Content: "gofmt"})
	got := ids(t, idx, Query{Q: "gofmt", Section: "review"})
	if len(got) != 1 || got[0] != "global/review/a" {
		t.Errorf("hits = %v, want only review/a", got)
	}
}
