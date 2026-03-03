package handler

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sort"

	"gogamemaps/internal/db"
	"gogamemaps/internal/models"
	"gogamemaps/internal/similar"

	"github.com/gorilla/mux"
)

type mapRec struct {
	AppID     int64
	Name      string
	Score     float64
	AnchorApp int64
}

func (s *Server) handleGetUserMap(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := readPathInt64(vars, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Payload{Status: "error", Error: "Invalid user ID"})
		return
	}
	if !s.requireUserAccess(w, r, userID) {
		return
	}

	user, err := db.GetUserByID(r.Context(), s.DB, userID)
	if err != nil {
		writeServerError(w, "Failed to load user map.", err)
		return
	}
	if user == nil {
		writeJSON(w, http.StatusNotFound, Payload{Status: "error", Error: "User not found"})
		return
	}

	payload, err := s.buildUserMapPayload(r.Context(), user)
	if err != nil {
		writeServerError(w, "Failed to build user map.", err)
		return
	}

	writeJSON(w, http.StatusOK, Payload{Status: "success", Data: payload})
}

func (s *Server) buildUserMapPayload(ctx context.Context, user *models.User) (models.MapPayload, error) {
	payload := models.MapPayload{UserID: user.ID}

	userNodeID := fmt.Sprintf("user-%d", user.ID)
	nodes := make([]models.MapNode, 0, 1+len(user.GamesLiked)+12)
	edges := make([]models.MapEdge, 0, len(user.GamesLiked)+12)
	nodes = append(nodes, models.MapNode{
		ID:    userNodeID,
		Label: user.Name,
		Kind:  "user",
		X:     0,
		Y:     0,
	})

	if len(user.GamesLiked) == 0 {
		payload.Nodes = nodes
		payload.Edges = edges
		return payload, nil
	}

	names, err := db.GetGameNamesByAppIDs(ctx, s.DB, user.GamesLiked)
	if err != nil {
		return payload, err
	}

	likedAngles := make(map[int64]float64, len(user.GamesLiked))
	likedEmbeddings := make(map[int64]map[string]float64, len(user.GamesLiked))
	likedRadius := 160.0
	totalLiked := float64(len(user.GamesLiked))

	for i, appID := range user.GamesLiked {
		angle := 2 * math.Pi * (float64(i) / totalLiked)
		likedAngles[appID] = angle
		x := math.Cos(angle) * likedRadius
		y := math.Sin(angle) * likedRadius
		label := names[appID]
		if label == "" {
			label = fmt.Sprintf("Game %d", appID)
		}
		nodeID := fmt.Sprintf("game-%d", appID)
		nodes = append(nodes, models.MapNode{
			ID:    nodeID,
			Label: label,
			Kind:  "liked",
			AppID: appID,
			X:     x,
			Y:     y,
		})
		edges = append(edges, models.MapEdge{
			From: userNodeID,
			To:   nodeID,
			Kind: "liked",
		})

		emb, err := db.GetGameEmbeddingByAppID(ctx, s.DB, appID)
		if err == nil && len(emb) > 0 {
			likedEmbeddings[appID] = emb
		}
	}

	if len(user.TasteEmbedding) == 0 {
		payload.Nodes = nodes
		payload.Edges = edges
		return payload, nil
	}

	recommendations, err := similar.FindGamesForUserTaste(ctx, s.DB, user.ID, 12)
	if err != nil {
		return payload, err
	}

	recGroups := make(map[int64][]mapRec, len(user.GamesLiked))
	for _, rec := range recommendations {
		anchorID := user.GamesLiked[0]
		if len(likedEmbeddings) > 0 {
			recEmb, err := db.GetGameEmbeddingByAppID(ctx, s.DB, rec.AppID)
			if err == nil && len(recEmb) > 0 {
				bestScore := -1.0
				for likedID, likedEmb := range likedEmbeddings {
					score := similar.CosineSim(recEmb, likedEmb)
					if score > bestScore {
						bestScore = score
						anchorID = likedID
					}
				}
			}
		}
		recGroups[anchorID] = append(recGroups[anchorID], mapRec{
			AppID:     rec.AppID,
			Name:      rec.Name,
			Score:     rec.Score,
			AnchorApp: anchorID,
		})
	}

	recBaseRadius := 280.0
	recSpread := 140.0

	for anchorID, list := range recGroups {
		sort.Slice(list, func(i, j int) bool { return list[i].Score > list[j].Score })
		n := len(list)
		spacing := 0.35
		if n > 1 {
			spacing = math.Min(0.35, 1.2/float64(n))
		}
		anchorAngle := likedAngles[anchorID]
		anchorNodeID := fmt.Sprintf("game-%d", anchorID)
		for idx, rec := range list {
			offset := (float64(idx) - float64(n-1)/2) * spacing
			angle := anchorAngle + offset
			score := clamp01(rec.Score)
			radius := recBaseRadius + (1.0-score)*recSpread
			x := math.Cos(angle) * radius
			y := math.Sin(angle) * radius

			recNodeID := fmt.Sprintf("rec-%d", rec.AppID)
			nodes = append(nodes, models.MapNode{
				ID:     recNodeID,
				Label:  rec.Name,
				Kind:   "recommended",
				AppID:  rec.AppID,
				X:      x,
				Y:      y,
				Score:  rec.Score,
				Anchor: anchorNodeID,
			})
			edges = append(edges, models.MapEdge{
				From: anchorNodeID,
				To:   recNodeID,
				Kind: "recommended",
			})
		}
	}

	payload.Nodes = nodes
	payload.Edges = edges
	return payload, nil
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
