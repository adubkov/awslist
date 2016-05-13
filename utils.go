package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

var (
// @readonly
)

// Convert []*string to string.
// If fills nil values with "None"
func makeFormattedOutput(i []*string) string {
	s := []string{}
	for _, str := range i {
		if str == nil {
			s = append(s, "None")
		} else {
			s = append(s, *str)
		}
	}

	i_parts := strings.Join(s, ",")
	return fmt.Sprintf("%s\n", i_parts)
}

func printJson(res io.Writer, v interface{}) {
	data, _ := json.MarshalIndent(v, "", "\t")
	fmt.Fprintf(res, string(data[:]))
}

func printText(res io.Writer, s string) {
	fmt.Fprintf(res, s)
}
