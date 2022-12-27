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
		t.Run(entry.Name(), func(t *testing.T) {
			in, err := os.ReadFile(filepath.Join("testdata", entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(filepath.Join("testdata", strings.TrimSuffix(entry.Name(), ".md")+".html"))
			if err != nil {
				t.Fatal(err)
			}
			got := bytes.Buffer{}
			extender := &Extender{}
			cfgfile, err := os.ReadFile(filepath.Join("testdata", strings.TrimSuffix(entry.Name(), ".md")+".json"))
			if !os.IsNotExist(err) {
				err = json.Unmarshal(cfgfile, &extender)
				if err != nil {
					t.Fatal(err)
				}
			}
			err = goldmark.New(goldmark.WithExtensions(extender)).Convert(in, &got)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(want, got.Bytes()); diff != "" {
				t.Fatalf("%s:\n\nwant:\n%s\n\ngot:\n%s\n\ndiff:\n%s\n", entry.Name(), want, got.String(), diff)
			}
		})
	}
}
