//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc_test

import (
	"bytes"
	"os"
	"slices"
	"testing"

	"github.com/net-of-arc/marc"
)

func benchmarkNode(b *testing.B) ([]byte, marc.Node) {
	b.Helper()
	src, err := os.ReadFile("testdata/bench-100.md")
	if err != nil {
		b.Fatal(err)
	}
	n, err := marc.Unmarshal(src)
	if err != nil {
		b.Fatal(err)
	}
	out, err := marc.Marshal(n)
	if err != nil || !bytes.Equal(out, src) || len(slices.Collect(n.Triples())) != 100 {
		b.Fatal("testdata/bench-100.md must be a canonical node of 100 statements")
	}
	return src, n
}

func BenchmarkDecode(b *testing.B) {
	src, _ := benchmarkNode(b)
	b.ReportAllocs()
	b.SetBytes(int64(len(src)))
	for b.Loop() {
		if _, err := marc.Unmarshal(src); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	src, n := benchmarkNode(b)
	b.ReportAllocs()
	b.SetBytes(int64(len(src)))
	for b.Loop() {
		if _, err := marc.Marshal(n); err != nil {
			b.Fatal(err)
		}
	}
}
