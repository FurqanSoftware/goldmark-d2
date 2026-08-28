package d2

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
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

			wantN, gotN := normalize(want), normalize(got.Bytes())
			if diff := cmp.Diff(wantN, gotN); diff != "" {
				t.Fatalf("%s:\n\nwant:\n%s\n\ngot:\n%s\n\ndiff:\n%s\n", entry.Name(), wantN, gotN, diff)
			}
		})
	}
}

// fontDataURI matches the payload of a base64 data URI, which D2 uses to embed
// fonts in the SVG it renders.
var fontDataURI = regexp.MustCompile(`(base64,)[A-Za-z0-9+/]+=*`)

// normalize blanks out embedded font payloads. D2 compresses fonts with
// compress/zlib, whose output varies between Go releases, so comparing those
// bytes would tie the golden files to the toolchain that generated them. The
// payload is still required to be present and non-empty; everything else,
// including all geometry, is compared as-is.
func normalize(b []byte) []byte {
	return fontDataURI.ReplaceAll(b, []byte("${1}<font>"))
}
