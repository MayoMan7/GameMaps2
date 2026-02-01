package tfidf

import (
	"fmt"
	"strings"
	"unicode"

	"gogamemaps/internal/models"
)

var stopwords = map[string]struct{}{
	"the": {}, "and": {}, "of": {}, "to": {}, "in": {}, "as": {}, "at": {}, "up": {},
	"for": {}, "a": {}, "an": {}, "on": {}, "is": {}, "it": {}, "this": {}, "that": {},
	"with": {}, "by": {}, "or": {}, "from": {}, "be": {}, "also": {}, "well": {},
}

// TokenizeString tokenizes a string into lowercase tokens, filtering short and stopwords.
func TokenizeString(s string) []string {
	s = strings.ToLower(s)
	var tokens []string
	var b strings.Builder
	flush := func() {
		if b.Len() == 0 {
			return
		}
		tok := b.String()
		b.Reset()
		if len(tok) < 2 {
			return
		}
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

func tokenizeArrayOfString(arr []string) []string {
	var tokens []string
	for _, s := range arr {
		tokens = append(tokens, TokenizeString(s)...)
	}
	return tokens
}

func tokenizeMapStringInt(m map[string]int) []string {
	var tokens []string
	for k := range m {
		tokens = append(tokens, TokenizeString(k)...)
	}
	return tokens
}

// TokenizeGame tokenizes a game into a document for TF-IDF.
func TokenizeGame(game *models.Game) []string {
	var tokens []string
	tokens = append(tokens, TokenizeString(game.Name)...)
	tokens = append(tokens, TokenizeString(game.ShortDescription)...)
	tokens = append(tokens, TokenizeString(game.DetailedDescription)...)
	tokens = append(tokens, TokenizeString(game.AboutTheGame)...)
	tokens = append(tokens, tokenizeArrayOfString(game.Developers)...)
	tokens = append(tokens, tokenizeArrayOfString(game.Publishers)...)
	tokens = append(tokens, tokenizeArrayOfString(game.Categories)...)
	tokens = append(tokens, tokenizeArrayOfString(game.Genres)...)
	tokens = append(tokens, tokenizeMapStringInt(game.Tags)...)
	return tokens
}

// TokenizeInt returns a single string token for an int.
func TokenizeInt(n int) []string {
	return []string{fmt.Sprintf("%d", n)}
}
