package main

import (
	"bytes"
	"testing"
)

func Test_makeFormattedOutput(t *testing.T) {
	s1, s2 := "s1", "s2"
	s := []*string{&s1, nil, &s2}
	result := "s1,None,s2\n"

	v := makeFormattedOutput(s)
	if v != result {
		t.Error("Excepcted:", result, " Got:", v)
	}
}

func Test_printJson(t *testing.T) {
	o := "o"
	s := []*string{&o}
	result := "[\n\t\"o\"\n]"

	var b bytes.Buffer
	printJson(&b, s)
	v := b.String()
	if v != result {
		t.Error("Excepcted:", result, " Got:", v)
	}
}

func Test_printText(t *testing.T) {
	s := "test"
	result := "test"

	var b bytes.Buffer
	printText(&b, s)
	v := b.String()
	if v != result {
		t.Error("Excepcted:", result, " Got:", v)
	}
}

func Test_jsonMarshal(t *testing.T) {
	o := "o"
	s := []*string{&o}
	result := `["o"]`

	v := jsonMarshal(s)
	if v != result {
		t.Error("Excepcted:", result, " Got:", v)
	}
}

func Test_strReplace(t *testing.T) {
	s := "[t{e\"s}t]"
	charset := "[]\"'{}"
	result := "test"

	v := strReplace(s, charset, "")
	if v != result {
		t.Error("Excepcted:", result, " Got:", v)
	}
}

func Test_formatSliceOutput(t *testing.T) {
	s1, s2 := "s1", "s2"
	s := []*string{&s1, &s2}
	result := "s1 s2"

	v := formatSliceOutput(s)
	if v != result {
		t.Error("Excepcted:", result, " Got:", v)
	}
}
