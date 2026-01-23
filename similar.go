// similar.go
// Minimal helpers to find similar games using TF-IDF embeddings stored as map[string]float64.
// Assumes you already computed embeddings per game with: map[token]tfidfScore.

package main

import (
	"math"
	"sort"
)

// Cosine similarity between two sparse vectors represented as map[token]weight.
func cosineSim(a, b map[string]float64) float64 {
	// iterate over smaller map for speed
	if len(a) > len(b) {
		a, b = b, a
	}

	var dot, na, nb float64

	for k, av := range a {
		dot += av * b[k]
		na += av * av
	}
	for _, bv := range b {
		nb += bv * bv
	}

	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

type SimilarResult struct {
	Index int
	AppID int64
	Name  string
	Score float64
}

// FindSimilarGames returns the topK most similar games to games[targetIdx].
// embeddings[i] must correspond to games[i].
func FindSimilarGames(games []Game, embeddings []map[string]float64, targetIdx int, topK int) []SimilarResult {
	if targetIdx < 0 || targetIdx >= len(games) || targetIdx >= len(embeddings) {
		return nil
	}
	if topK <= 0 {
		return nil
	}

	target := embeddings[targetIdx]
	results := make([]SimilarResult, 0, topK+16)

	for i := range games {
		if i == targetIdx || i >= len(embeddings) {
			continue
		}
		score := cosineSim(target, embeddings[i])
		if score <= 0 {
			continue
		}
		results = append(results, SimilarResult{
			Index: i,
			AppID: games[i].AppID,
			Name:  games[i].Name,
			Score: score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > topK {
		results = results[:topK]
	}
	return results
}

// SharedTopTerms returns up to topN overlapping tokens ranked by contribution.
// Great for explaining *why* two games are similar.
func SharedTopTerms(a, b map[string]float64, topN int) []struct {
	Term  string
	Score float64
} {
	if topN <= 0 {
		return nil
	}

	// iterate smaller map
	if len(a) > len(b) {
		a, b = b, a
	}

	type termScore struct {
		Term  string
		Score float64
	}

	list := make([]termScore, 0, topN+16)
	for term, av := range a {
		if bv, ok := b[term]; ok {
			// contribution proxy: min weight
			s := av
			if bv < s {
				s = bv
			}
			if s > 0 {
				list = append(list, termScore{Term: term, Score: s})
			}
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Score > list[j].Score
	})

	if len(list) > topN {
		list = list[:topN]
	}

	// return anonymous struct slice (simple printing)
	out := make([]struct {
		Term  string
		Score float64
	}, len(list))

	for i := range list {
		out[i] = struct {
			Term  string
			Score float64
		}{Term: list[i].Term, Score: list[i].Score}
	}

	return out
}
