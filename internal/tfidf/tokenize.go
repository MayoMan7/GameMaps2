package tfidf

import (
	"fmt"
	"math"
	"strings"
	"unicode"

	"gogamemaps/internal/models"
)

var stopwords = map[string]struct{}{
	"the": {}, "and": {}, "of": {}, "to": {}, "in": {}, "as": {}, "at": {}, "up": {},
	"for": {}, "a": {}, "an": {}, "on": {}, "is": {}, "it": {}, "this": {}, "that": {},
	"with": {}, "by": {}, "or": {}, "from": {}, "be": {}, "also": {}, "well": {},
}

var franchiseNoise = map[string]struct{}{
	"edition": {}, "definitive": {}, "ultimate": {}, "complete": {}, "goty": {}, "remastered": {},
	"remaster": {}, "collection": {}, "bundle": {}, "pack": {}, "dlc": {}, "season": {},
	"episode": {}, "chapter": {}, "anniversary": {}, "enhanced": {}, "expansion": {},
	"beta": {}, "demo": {}, "soundtrack": {}, "upgrade": {}, "deluxe": {},
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

func isNumericToken(tok string) bool {
	for _, r := range tok {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return tok != ""
}

func isRomanNumeral(tok string) bool {
	for _, r := range tok {
		switch r {
		case 'i', 'v', 'x', 'l', 'c', 'd', 'm':
		default:
			return false
		}
	}
	return tok != ""
}

func applyTokenWeight(tok string, base float64) float64 {
	if base <= 0 {
		return 0
	}
	if _, ok := franchiseNoise[tok]; ok {
		base *= 0.2
	}
	if isNumericToken(tok) || isRomanNumeral(tok) {
		base *= 0.2
	}
	return base
}

func addWeightedTokens(out map[string]float64, tokens []string, weight float64) {
	if weight <= 0 {
		return
	}
	for _, tok := range tokens {
		w := applyTokenWeight(tok, weight)
		if w <= 0 {
			continue
		}
		out[tok] += w
	}
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

// TokenizeGameWeighted tokenizes a game into weighted tokens for TF-IDF.
func TokenizeGameWeighted(game *models.Game) map[string]float64 {
	out := make(map[string]float64, 256)

	// Field weights: reduce title dominance, favor tags/genres for experience similarity.
	addWeightedTokens(out, TokenizeString(game.Name), 0.1)
	addWeightedTokens(out, TokenizeString(game.ShortDescription), 0.6)
	addWeightedTokens(out, TokenizeString(game.DetailedDescription), 0.25)
	addWeightedTokens(out, TokenizeString(game.AboutTheGame), 0.35)
	addWeightedTokens(out, tokenizeArrayOfString(game.Developers), 0.25)
	addWeightedTokens(out, tokenizeArrayOfString(game.Publishers), 0.2)
	addWeightedTokens(out, tokenizeArrayOfString(game.Categories), 0.9)
	addWeightedTokens(out, tokenizeArrayOfString(game.Genres), 1.1)

	for tag, count := range game.Tags {
		if tag == "" {
			continue
		}
		toks := TokenizeString(tag)
		if len(toks) == 0 {
			continue
		}
		boost := math.Log1p(float64(count))
		if boost < 1 {
			boost = 1
		}
		addWeightedTokens(out, toks, 1*boost)
	}

	return out
}

// TokenizeInt returns a single string token for an int.
func TokenizeInt(n int) []string {
	return []string{fmt.Sprintf("%d", n)}
}
