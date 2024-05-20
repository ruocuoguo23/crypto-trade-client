package rpc

import (
	"strings"
	"unicode"

	"crypto-trade-client/common/stringutil"
	"crypto-trade-client/common/web/fetch"
)

// Naming convention https://en.wikipedia.org/wiki/Naming_convention_(programming)#Multiple-word_identifiers
const (
	CamelCase  NamingConvention = iota // CamelCase format is `twoWords`
	PascalCase                         // PascalCase format is `TwoWords`
	SnakeCase                          // SnakeCase format is `two_words`
	LowerCase                          // LowerCase format is `towwords`
	Original
)

type NamingConvention uint

type methodName interface {
	MethodNamingConvention() NamingConvention
}

type namespace interface {
	Namespace() string
	NamespaceSeparator() string
}

type middleware interface {
	BeforeRequest() []fetch.RequestMiddleware
}

// CamelCaseName convert the name to lower camel case format.
// CamelCaseName regex pattern is [a-z]+((\d)|([A-Z0-9][a-z0-9]+))*([A-Z])?,
// includes letters and numbers.
func CamelCaseName(name string) string {
	words := stringutil.SplitWordsByCamelCase(name)
	if len(words) == 0 {
		return name
	}
	r := []rune(words[0])
	r[0] = unicode.ToLower(r[0])
	words[0] = string(r)
	return strings.Join(words, "")
}

func LowerCaseName(name string) string {
	return strings.ToLower(name)
}

func SnakeCaseName(name string) string {
	words := stringutil.SplitWordsByCamelCase(name)
	if len(words) == 0 {
		return name
	}

	for i := 0; i < len(words); i++ {
		words[i] = strings.ToLower(words[i])
	}

	return strings.Join(words, "_")
}
