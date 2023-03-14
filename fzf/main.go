package fzf

import (
	"github.com/junegunn/fzf/src/algo"
	"github.com/junegunn/fzf/src/util"
)

func Score(input, pattern string) int {
	chars := util.ToChars([]byte(input))
	res, _ := algo.FuzzyMatchV2(false, false, true, &chars, []rune(pattern), false, nil)

	return res.Score
}
