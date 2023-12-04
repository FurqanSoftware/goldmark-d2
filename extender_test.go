package d2

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yuin/goldmark"
)

func TestExtenderTestExtender(t *testing.T) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".md")
		t.Run(entry.Name(), func(t *testing.T) {
			in, err := os.ReadFile(filepath.Join("testdata", entry.Name()))
			if err != nil {
				t.Fatal(err)
			}

			want, wantErr := os.ReadFile(filepath.Join("testdata", name+".html"))
			if wantErr != nil {
				t.Error(wantErr)
			}

			extender := &Extender{}
			cfgfile, err := os.ReadFile(filepath.Join("testdata", name+".json"))
			if !os.IsNotExist(err) {
				err = json.Unmarshal(cfgfile, &extender)
				if err != nil {
					t.Fatal(err)
				}
			}

			got := bytes.Buffer{}
			err = goldmark.New(goldmark.WithExtensions(extender)).Convert(in, &got)
			if err != nil {
				t.Fatal(err)
			}

			if os.IsNotExist(wantErr) {
				if err := os.WriteFile(
					filepath.Join("testdata", name+".html"),
					got.Bytes(),
					0666,
				); err != nil {
					t.Fatal(err)
				}
			}

			if diff := cmp.Diff(want, got.Bytes()); diff != "" {
				t.Fatalf("%s:\n\nwant:\n%s\n\ngot:\n%s\n\ndiff:\n%s\n", entry.Name(), want, got.String(), diff)
			}
		})
	}
}
