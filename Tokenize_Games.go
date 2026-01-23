package main

import (
	"fmt"
	"strings"
	"unicode"
)

var stopwords = map[string]struct{}{
	"the": {}, "and": {}, "of": {}, "to": {}, "in": {}, "as": {}, "at": {}, "up": {},
	"for": {}, "a": {}, "an": {}, "on": {}, "is": {}, "it": {}, "this": {}, "that": {},
	"with": {}, "by": {}, "or": {}, "from": {}, "be": {}, "also": {}, "well": {},
}

func tokenizeString(description string) []string {
	// 1) normalize
	s := strings.ToLower(description)

	// 2) split into tokens by keeping only letters/digits
	var tokens []string
	var b strings.Builder

	flush := func() {
		if b.Len() == 0 {
			return
		}
		tok := b.String()
		b.Reset()

		// optional: filter short tokens
		if len(tok) < 2 {
			return
		}
		// optional: stopword removal
		if _, ok := stopwords[tok]; ok {
			return
		}
		tokens = append(tokens, tok)
	}

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()

	return tokens
}

func tokenizeInt(n int) []string {
	return []string{fmt.Sprintf("%d", n)}
}

func tokenizeArrayofString(arr []string) []string {
	var tokens []string
	for _, s := range arr {
		tokens = append(tokens, tokenizeString(s)...)
	}
	return tokens
}

func tokenizeMapStringInt(m map[string]int) []string {
	var tokens []string
	for k, _ := range m {
		tokens = append(tokens, tokenizeString(k)...)
	}
	return tokens
}

func tokenizeGame(game *Game) []string {
	var tokens []string
	tokens = append(tokens, tokenizeString(game.Name)...)
	tokens = append(tokens, tokenizeString(game.ShortDescription)...)
	tokens = append(tokens, tokenizeString(game.DetailedDescription)...)
	tokens = append(tokens, tokenizeString(game.AboutTheGame)...)
	tokens = append(tokens, tokenizeArrayofString(game.Developers)...)
	tokens = append(tokens, tokenizeArrayofString(game.Publishers)...)
	tokens = append(tokens, tokenizeArrayofString(game.Categories)...)
	tokens = append(tokens, tokenizeArrayofString(game.Genres)...)
	tokens = append(tokens, tokenizeMapStringInt(game.Tags)...)
	return tokens
}
