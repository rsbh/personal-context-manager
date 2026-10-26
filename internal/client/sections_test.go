package client

import (
	"context"
	"net/http"
	"reflect"
	"testing"
)

func TestListSections(t *testing.T) {
	f := &fakeAPI{status: http.StatusOK, resp: `["review","style"]`}
	got, err := f.start(t).ListSections(context.Background(), "project", "api")
	if err != nil {
		t.Fatal(err)
	}
	f.expect(t, http.MethodGet, "/api/v1/sections", "project=api&scope=project", "")
	if want := []string{"review", "style"}; !reflect.DeepEqual(got, want) {
		t.Errorf("sections = %v, want %v", got, want)
	}
}

func TestListSectionsGlobal(t *testing.T) {
	f := &fakeAPI{status: http.StatusOK, resp: `[]`}
	if _, err := f.start(t).ListSections(context.Background(), "global", ""); err != nil {
		t.Fatal(err)
	}
	f.expect(t, http.MethodGet, "/api/v1/sections", "scope=global", "")
}

func TestCreateSection(t *testing.T) {
	f := &fakeAPI{status: http.StatusCreated, resp: `{}`}
	if err := f.start(t).CreateSection(context.Background(), "project", "api", "style"); err != nil {
		t.Fatal(err)
	}
	f.expect(t, http.MethodPost, "/api/v1/sections", "", `{"scope":"project","project":"api","name":"style"}`+"\n")
}

func TestDeleteSection(t *testing.T) {
	f := &fakeAPI{status: http.StatusNoContent}
	if err := f.start(t).DeleteSection(context.Background(), "global", "", "style", true); err != nil {
		t.Fatal(err)
	}
	f.expect(t, http.MethodDelete, "/api/v1/sections/style", "force=true&scope=global", "")
}
