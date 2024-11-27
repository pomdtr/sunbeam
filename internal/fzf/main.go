package fzf

import (
	"unicode"

	"github.com/junegunn/fzf/src/algo"
	"github.com/junegunn/fzf/src/util"
)

func IsLower(s string) bool {
	for _, r := range s {
		if !unicode.IsLower(r) && unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func Score(input, pattern string) int {
	chars := util.ToChars([]byte(input))

	// follow the out of the box logic of the fzf CLI, making the search "smart-case" sensitive
	// if there is an upper case letter the search is case sensitive, otherwise it is not
	caseSensitive := !IsLower(pattern)
	res, _ := algo.FuzzyMatchV1(caseSensitive, false, true, &chars, []rune(pattern), false, nil)

	return res.Score
}
