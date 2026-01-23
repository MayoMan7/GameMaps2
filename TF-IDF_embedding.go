package main

import "math"

// term = individual word
// document = []string
// corpus = [][]string

// preComputeSeenTerms returns the unique vocab in the corpus.
// (Fast, but you don't actually need this if you compute DF; df keys == vocab.)
func preComputeSeenTerms(corpus [][]string) map[string]struct{} {
	seenTerms := make(map[string]struct{}, 1<<16)
	for _, doc := range corpus {
		for _, term := range doc {
			seenTerms[term] = struct{}{}
		}
	}
	return seenTerms
}

// precomputeDocumentsContainingTerm computes DF correctly:
// df[term] = number of documents that contain term (counted once per document).
func precomputeDocumentsContainingTerm(corpus [][]string) map[string]int {
	df := make(map[string]int, 1<<16)

	for _, doc := range corpus {
		// avoid counting the same term multiple times in a doc
		seen := make(map[string]struct{}, len(doc))
		for _, term := range doc {
			if _, ok := seen[term]; ok {
				continue
			}
			seen[term] = struct{}{}
			df[term]++
		}
	}

	return df
}

// PrecmputeIDF computes idf from DF. You do NOT need seenTerms; df keys are the terms.
// This version also uses a smoothed IDF (more stable).
func PrecmputeIDF(corpus [][]string, df map[string]int) map[string]float64 {
	N := float64(len(corpus))
	idf := make(map[string]float64, len(df))

	// Smoothed IDF: ln((1+N)/(1+df)) + 1
	// Better than ln(N/df) because it avoids weird extremes.
	for term, dfi := range df {
		idf[term] = math.Log((1.0+N)/(1.0+float64(dfi))) + 1.0
	}

	return idf
}

// TFIDFEmbedding builds a sparse tf-idf vector for a single doc.
// Optimizations:
// - uses int counts (cheaper than float increments)
// - prealloc maps based on doc size
// - avoids repeated float64(len(document)) conversions
func TFIDFEmbedding(document []string, idf map[string]float64) map[string]float64 {
	if len(document) == 0 {
		return map[string]float64{}
	}

	// Count tokens once
	counts := make(map[string]int, len(document))
	for _, tok := range document {
		counts[tok]++
	}

	den := float64(len(document))

	// Build embedding (sparse)
	emb := make(map[string]float64, len(counts))
	for tok, c := range counts {
		// If a token isn't in idf (e.g. built IDF on a cutoff), treat its idf as 0 and skip.
		idfv, ok := idf[tok]
		if !ok || idfv == 0 {
			continue
		}
		tf := float64(c) / den
		emb[tok] = tf * idfv
	}

	return emb
}
