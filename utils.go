package main

import (
	"fmt"
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