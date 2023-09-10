package fzf

import (
	"strings"

	"github.com/junegunn/fzf/src/algo"
	"github.com/junegunn/fzf/src/util"
)

func Score(input, pattern string) int {
	chars := util.ToChars([]byte(input))
	res, _ := algo.FuzzyMatchV2(false, false, true, &chars, []rune(strings.ToLower(pattern)), false, nil)

	return res.Score
}
