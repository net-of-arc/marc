//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/net-of-arc/marc"
)

func seed(f *testing.F) {
	err := filepath.WalkDir("testdata", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}
		b, err := os.ReadFile(path)
		f.Add(b)
		return err
	})
	if err != nil {
		f.Fatal(err)
	}
}

func FuzzDecode(f *testing.F) {
	seed(f)
	f.Fuzz(func(t *testing.T, b []byte) {
		n, err := marc.Unmarshal(b)
		if err == nil && n.ID().IsZero() {
			t.Fatal("decoded a node without identity")
		}
	})
}

func FuzzRoundTrip(f *testing.F) {
	seed(f)
	f.Fuzz(func(t *testing.T, b []byte) {
		n1, err := marc.Unmarshal(b)
		if err != nil {
			return
		}
		b1, err := marc.Marshal(n1)
		if errors.Is(err, marc.ErrProseStructure) {
			return
		}
		n2 := mustRoundTrip(t, b1, err)
		equalTriples(t, n1, n2)

		b2, err := marc.Marshal(n2)
		n3 := mustRoundTrip(t, b2, err)
		b3, err := marc.Marshal(n3)
		if err != nil || string(b3) != string(b2) {
			t.Fatalf("canonical form is not a fixpoint:\n%q\n%q", b2, b3)
		}
	})
}

func mustRoundTrip(t *testing.T, b []byte, err error) marc.Node {
	t.Helper()
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	n, err := marc.Unmarshal(b)
	if err != nil {
		t.Fatalf("re-decode failed: %v\n%q", err, b)
	}
	return n
}

// equalTriples compares two nodes' triples as a set (BUG-001: a decoded
// node's facts follow document order, which need not match the role-grouped
// order the encoder reconstructs when re-decoded).
func equalTriples(t *testing.T, a, b marc.Node) {
	t.Helper()
	ta, tb := sortedTriples(a), sortedTriples(b)
	if !reflect.DeepEqual(ta, tb) {
		t.Fatalf("triples differ:\n%v\n%v", ta, tb)
	}
}

func sortedTriples(n marc.Node) []marc.Triple {
	ts := slices.Collect(n.Triples())
	slices.SortFunc(ts, func(a, b marc.Triple) int {
		if c := strings.Compare(a.Predicate.String(), b.Predicate.String()); c != 0 {
			return c
		}
		return strings.Compare(a.Object.String(), b.Object.String())
	})
	return ts
}
