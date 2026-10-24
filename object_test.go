//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc

import (
	"errors"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/fogfish/it/v2"
)

func TestObjectEqualityKeys(t *testing.T) {
	at := time.Date(2026, 10, 23, 12, 0, 0, 0, time.FixedZone("CEST", 2*3600))
	for _, pair := range [][2]Object{
		{Text(" a "), Text("a")},
		{Text("Café"), Text("Café")},
		{Text("line  \nnext\t"), Text("line\nnext")},
		{RefTo(MustID("X")).WithDisplay("d"), RefTo(MustID("X"))},
		{DateTimeOf(at), DateTimeOf(at.UTC())},
		{DateOf(at), DateOf(time.Date(2026, 10, 23, 23, 0, 0, 0, time.UTC))},
	} {
		it.Then(t).Should(it.Equal(pair[0].key(), pair[1].key()))
	}

	number, _ := ParseNumber("1")
	it.Then(t).ShouldNot(it.Equal(Text("1").key(), number.key()))
}

func TestParseNumber(t *testing.T) {
	for _, s := range []string{"42", "-7", "1.5", "1e3", "0x1F", "0o17", ".inf", "-.inf", ".nan", "1_000"} {
		n, err := ParseNumber(s)
		it.Then(t).Should(it.Nil(err), it.Equal(n.String(), s))
	}
	for _, s := range []string{"abc", "", "1 2", "true", "2026-10-23", "1\n2"} {
		_, err := ParseNumber(s)
		it.Then(t).Should(it.True(errors.Is(err, ErrInvalidObject)))
	}
}

func TestNumberConversions(t *testing.T) {
	i, iok := mustNumber("42").Int64()
	f, fok := mustNumber("1.5").Float64()
	h, hok := mustNumber("0x1F").Float64()
	inf, infok := mustNumber("-.inf").Float64()
	_, nok := mustNumber("1.5").Int64()

	it.Then(t).Should(
		it.True(iok), it.Equal(i, int64(42)),
		it.True(fok), it.Equal(f, 1.5),
		it.True(hok), it.Equal(h, 31.0),
		it.True(infok), it.True(math.IsInf(inf, -1)),
		it.True(!nok),
	)
}

func mustNumber(s string) Number {
	n, err := ParseNumber(s)
	if err != nil {
		panic(err)
	}
	return n
}

func TestObjectString(t *testing.T) {
	for obj, s := range map[Object]string{
		RefTo(MustID("T")):                      "[[T]]",
		RefTo(MustID("T")).WithDisplay("shown"): "[[T|shown]]",
		Text("plain"):                           "plain",
		mustNumber("42"):                        "42",
		Bool(true):                              "true",
		Bool(false):                             "false",
		DateOf(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)):                "2026-01-02",
		DateTimeOf(time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)):         "2026-10-23T10:00:00Z",
		DateTimeOf(time.Date(2026, 10, 23, 10, 0, 0, 500000000, time.UTC)): "2026-10-23T10:00:00.5Z",
	} {
		it.Then(t).Should(it.Equal(obj.String(), s))
	}
}

func TestRefAccessors(t *testing.T) {
	r := RefTo(MustID("T")).WithDisplay("d")
	it.Then(t).Should(
		it.Equal(r.Target(), MustID("T")),
		it.Equal(r.Display(), "d"),
		it.Equal(RefTo(MustID("T")).Display(), ""),
	)
}

func TestDateAccessors(t *testing.T) {
	at := time.Date(2026, 10, 23, 22, 0, 0, 0, time.FixedZone("X", -5*3600))
	it.Then(t).Should(
		it.Equal(DateOf(at).Time(), time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC)),
		it.Equal(DateTimeOf(at).Time(), at.UTC()),
	)
}

func TestCanonicalOrder(t *testing.T) {
	early := DateOf(time.Date(999, 1, 1, 0, 0, 0, 0, time.UTC))
	late := DateOf(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	objects := []Object{
		Text("b"), RefTo(MustID("Z")), mustNumber("9"), Bool(true), late,
		DateTimeOf(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)), Text("a"),
		mustNumber("10"), early, Bool(false),
	}
	slices.SortFunc(objects, compareObjects)

	it.Then(t).Should(it.Seq(objects).Equal(
		Bool(false), Bool(true), early, late, mustNumber("10"), mustNumber("9"),
		RefTo(MustID("Z")), DateTimeOf(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
		Text("a"), Text("b"),
	))
}

func TestSelect(t *testing.T) {
	mixed := slices.Values([]Object{Text("a"), RefTo(MustID("X")), Bool(true), RefTo(MustID("Y"))})
	it.Then(t).Should(
		it.Seq(slices.Collect(Select[Ref](mixed))).Equal(RefTo(MustID("X")), RefTo(MustID("Y"))),
		it.Seq(slices.Collect(Select[Text](mixed))).Equal(Text("a")),
	)
}
