// Copyright (c) 2017 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

// Package path contains methods for dealing with key.Paths.
package path

import (
	"strings"

	"github.com/aristanetworks/goarista/key"
)

// New constructs a path from a variable number of elements.
// Each element may either be a key.Key or a value that can
// be wrapped by a key.Key.
func New(elements ...interface{}) key.Path {
	result := make(key.Path, len(elements))
	copyElements(result, elements...)
	return result
}

// Append appends a variable number of elements to a path.
// Each element may either be a key.Key or a value that can
// be wrapped by a key.Key. Note that calling Append on a
// single path returns that same path, whereas in all other
// cases a new path is returned.
func Append(path key.Path, elements ...interface{}) key.Path {
	if len(elements) == 0 {
		return path
	}
	n := len(path)
	result := make(key.Path, n+len(elements))
	copy(result, path)
	copyElements(result[n:], elements...)
	return result
}

// Join joins a variable number of paths together. Each path
// in the joining is treated as a subpath of its predecessor.
// Calling Join with no or only empty paths returns nil.
func Join(paths ...key.Path) key.Path {
	n := 0
	for _, path := range paths {
		n += len(path)
	}
	if n == 0 {
		return nil
	}
	result, i := make(key.Path, n), 0
	for _, path := range paths {
		i += copy(result[i:], path)
	}
	return result
}

// Parent returns all but the last element of the path. If
// the path is empty, Parent returns nil.
func Parent(path key.Path) key.Path {
	if len(path) > 0 {
		return path[:len(path)-1]
	}
	return nil
}

// Base returns the last element of the path. If the path is
// empty, Base returns nil.
func Base(path key.Path) key.Key {
	if len(path) > 0 {
		return path[len(path)-1]
	}
	return nil
}

// Clone returns a new path with the same elements as in the
// provided path.
func Clone(path key.Path) key.Path {
	result := make(key.Path, len(path))
	copy(result, path)
	return result
}

// Equal returns whether path a and path b are the same
// length and whether each element in b corresponds to the
// same element in a.
func Equal(a, b key.Path) bool {
	return len(a) == len(b) && hasPrefix(a, b)
}

// HasElement returns whether element b exists in path a.
func HasElement(a key.Path, b key.Key) bool {
	for _, element := range a {
		if element.Equal(b) {
			return true
		}
	}
	return false
}

// HasPrefix returns whether path b is a prefix of path a.
// It checks that b is at most the length of path a and
// whether each element in b corresponds to the same element
// in a from the first element.
func HasPrefix(a, b key.Path) bool {
	return len(a) >= len(b) && hasPrefix(a, b)
}

// HasPrefixString returns whether path b is a prefix of path a.
// It checks that b is at most the length of path a and
// that every element of a matches corresponding element of b exactly
// except the last element.
// The last element of a is considered a match if it matches exactly
// or if it is a string and the last element of a is a substring
// of corresponding element of b
func HasPrefixString(a, b key.Path) bool {
	return len(a) >= len(b) && hasPrefixString(a, b)
}

// Match returns whether path a and path b are the same
// length and whether each element in b corresponds to the
// same element or a wildcard in a.
func Match(a, b key.Path) bool {
	return len(a) == len(b) && matchPrefix(a, b)
}

// MatchPrefix returns whether path b is a prefix of path a
// where path a may contain wildcards.
// It checks that b is at most the length of path a and
// whether each element in b corresponds to the same element
// or a wildcard in a from the first element.
func MatchPrefix(a, b key.Path) bool {
	return len(a) >= len(b) && matchPrefix(a, b)
}

// MatchPrefixString returns whether path b is a prefix of path a
// where path a may contain wildcards.
// It checks that b is at most the length of path a and
// that every element of a matches corresponding element of b exactly
// or a wildcard except the last element.
// The last element of a is considered a match if it matches exactly
// or it is a wildcard or if it is a string and the last element of a
// is a substring of corresponding element of b
func MatchPrefixString(a, b key.Path) bool {
	return len(a) >= len(b) && matchPrefixString(a, b)
}

// FromString constructs a path from the elements resulting
// from a split of the input string by "/". Strings that do
// not lead with a '/' are accepted but not reconstructable
// with key.Path.String. Both "" and "/" are treated as a
// key.Path{}.
func FromString(str string) key.Path {
	if str == "" || str == "/" {
		return key.Path{}
	} else if str[0] == '/' {
		str = str[1:]
	}
	elements := strings.Split(str, "/")
	result := make(key.Path, len(elements))
	for i, element := range elements {
		result[i] = key.New(element)
	}
	return result
}

func copyElements(dest key.Path, elements ...interface{}) {
	for i, element := range elements {
		switch val := element.(type) {
		case key.Key:
			dest[i] = val
		default:
			dest[i] = key.New(val)
		}
	}
}

func hasPrefix(a, b key.Path) bool {
	for i := range b {
		if !b[i].Equal(a[i]) {
			return false
		}
	}
	return true
}

func compareElementString(a, b key.Key) bool {
	// If the element is a string, then it must be
	// a prefix of the corresponding element in a.
	bStr, bOk := b.Key().(string)
	aStr, aOk := a.Key().(string)
	if aOk && bOk {
		return strings.HasPrefix(aStr, bStr)
	}

	// Last element is not a string
	return b.Equal(a)
}

func hasPrefixString(a, b key.Path) bool {
	if len(b) == 0 {
		return true
	}

	if !hasPrefix(Parent(a), Parent(b)) {
		return false
	}

	// Compare the element in a that corresponds to last element of b.
	// This is needed because a can be longer than b.
	return compareElementString(a[len(b)-1], Base(b))
}

func matchPrefix(a, b key.Path) bool {
	for i := range b {
		if !a[i].Equal(Wildcard) && !b[i].Equal(a[i]) {
			return false
		}
	}
	return true
}

func matchPrefixString(a, b key.Path) bool {
	if len(b) == 0 {
		return true
	}

	if !matchPrefix(Parent(a), Parent(b)) {
		return false
	}

	// Compare the element in a that corresponds to last element of b.
	// This is needed because a can be longer than b.
	aKey := a[len(b)-1]

	return aKey.Equal(Wildcard) || compareElementString(aKey, Base(b))
}
