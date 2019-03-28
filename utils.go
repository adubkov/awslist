package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"io"
	"strings"
)

var (
// @readonly
)

func assumeRole(roleArn string) *credentials.Credentials {
	sess := session.Must(session.NewSession())
	return stscreds.NewCredentials(sess, roleArn)
}

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

func jsonMarshal(v interface{}) string {
	res, _ := json.Marshal(v)
	return string(res[:])
}

func strReplace(s, charset, r string) string {
	res := s
	for i := range charset {
		res = strings.Replace(res, string(charset[i]), r, -1)
	}
	return res
}

func formatSliceOutput(s []*string) string {
	chars := "[]\"'{}"
	res := jsonMarshal(s)
	res = strReplace(res, chars, "")
	res = strReplace(res, ",", " ")
	return res
}
