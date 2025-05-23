package sys

/*

  File:    lists.go
  Author:  Bob Shofner

  MIT License - https://opensource.org/license/mit/

  This permission notice shall be included in all copies
    or substantial portions of the Software.

*/
/*
  Description: Various []string functions
*/

// StringList is the type of array
type StringList []string

// Prepend appends a string to the beginning of a []string.
//
//	and optionally truncates to a maximum size
func Prepend(sl StringList, a string, max int) StringList {
	sl = append([]string{a}, sl...)
	if max > 0 && len(sl) > max {
		sl = sl[0:max]
	}
	return sl
}

// Add appends a string to the end of a []string.
func Add(sl StringList, a string, max int) StringList {
	sl = append([]string{a}, sl...)
	if len(sl) > max {
		sl = sl[0:max]
	}
	return sl
}

// Find finds a string using specified "equals" func.
func Find(sl StringList, eq func(int, string) bool) bool {
	for i, v := range sl {
		if eq(i, v) {
			return true
		}
	}
	return false
}

// Recent removes a string (if present) and prepends that string to the array.
// The maximum size of the list is preserved.
// For things like a limited list of "recent items"
func Recent(sl StringList, p string, max int) StringList {
	slr := Remove(sl, p)
	return Prepend(slr, p, max)
}

// Remove deletes a string from a []string.
func Remove(sl StringList, r string) StringList {
	for i, v := range sl {
		if v == r {
			return append(sl[:i], sl[i+1:]...)
		}
	}
	return sl
}

// Index finds the index of a string's index in an array. -1 == not found.
func Index(sl StringList, s string) int {
	for i, v := range sl {
		if v == s {
			return i
		}
	}
	return -1
}
