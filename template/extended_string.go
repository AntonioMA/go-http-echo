package template

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
)

// ExtendedString is a string that has some added methods to make it easier to use inside of a
// Template
type ExtendedString string

// Split implements the functionality of strings.Split. So Split
// slices ess into all substrings separated by sep and returns a slice of the substrings
// between those separators.
// If s does not contain sep and sep is not empty, Split returns a slice of length 1 whose only
// element is s.
// If sep is empty, Split splits after each UTF-8 sequence. If both s and sep are empty, Split
//  returns an empty slice.
func (es ExtendedString) Split(sep string) []ExtendedString {
	parts := strings.Split(string(es), sep)
	rv := make([]ExtendedString, len(parts))
	for i := 0; i < len(parts); i++ {
		rv[i] = ExtendedString(parts[i])
	}
	return rv
}

// LoadFile tries loading the file whose name is stored on es and returning the whole content of
// the file as an string
func (es ExtendedString) LoadFile() ExtendedString {
	if fileData, err := ioutil.ReadFile(string(es)); err != nil {
		return ""
	} else {
		return ExtendedString(fileData)
	}
}

// LoadRelativeFile tries loading the file whose name is stored on es, using basePath as the basePath
// (so es is assumed to be a relative path), and it returns returning the whole content of
// the file as an string
func (es ExtendedString) LoadRelativeFile(basePath string) ExtendedString {
	fullPath := strings.Join([]string{basePath, string(es)}, string(os.PathSeparator))
	if fileData, err := ioutil.ReadFile(fullPath); err != nil {
		return ""
	} else {
		return ExtendedString(fileData)
	}
}
func (es ExtendedString) LoadRelativeFileES(basePath ExtendedString) ExtendedString {
	return es.LoadRelativeFile(string(basePath))
}

// Fields implements the functionality of strings.Fields. So Fields
// splits the string s around each instance of one or more consecutive white space characters, as
// defined by unicode.IsSpace, returning a slice of substrings of s or an empty slice if s contains
// only white space.
func (es ExtendedString) Fields() []ExtendedString {
	fields := strings.Fields(string(es))
	rv := make([]ExtendedString, len(fields))
	for i := 0; i < len(fields); i++ {
		rv[i] = ExtendedString(fields[i])
	}
	return rv
}

// ToJSON returns the es string JSONified.
func (es ExtendedString) ToJSON() ExtendedString {
	if data, err := json.Marshal(es); err != nil {
		return ""
	} else {
		return ExtendedString(data)
	}
}

// ToBase64 returns the es string converted to Base64.
func (es ExtendedString) ToBase64() ExtendedString {
	return ExtendedString(base64.StdEncoding.EncodeToString([]byte(es)))
}
