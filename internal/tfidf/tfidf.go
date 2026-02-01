package tfidf

import "math"

// PrecomputeDocumentsContainingTerm computes DF: df[term] = number of docs containing term.
func PrecomputeDocumentsContainingTerm(corpus [][]string) map[string]int {
	df := make(map[string]int, 1<<16)
	for _, doc := range corpus {
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

// PrecomputeIDF computes IDF from DF. Smoothed: ln((1+N)/(1+df)) + 1.
func PrecomputeIDF(corpus [][]string, df map[string]int) map[string]float64 {
	N := float64(len(corpus))
	idf := make(map[string]float64, len(df))
	for term, dfi := range df {
		idf[term] = math.Log((1.0+N)/(1.0+float64(dfi))) + 1.0
	}
	return idf
}

// TFIDFEmbedding builds a sparse tf-idf vector for a single doc.
func TFIDFEmbedding(document []string, idf map[string]float64) map[string]float64 {
	if len(document) == 0 {
		return map[string]float64{}
	}
	counts := make(map[string]int, len(document))
	for _, tok := range document {
		counts[tok]++
	}
	den := float64(len(document))
	emb := make(map[string]float64, len(counts))
	for tok, c := range counts {
		idfv, ok := idf[tok]
		if !ok || idfv == 0 {
			continue
		}
		tf := float64(c) / den
		emb[tok] = tf * idfv
	}
	return emb
}
