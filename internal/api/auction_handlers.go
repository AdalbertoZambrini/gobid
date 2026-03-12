package api

import (
	"errors"
	"gobid/internal/jsonutils"
	"gobid/internal/services"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

func (api *Api) handleSubscribeUserToAuction(w http.ResponseWriter, r *http.Request) {
	rawProductId := strings.TrimSpace(chi.URLParam(r, "product_id"))
	if rawProductId == "" {
		trimmedPath := strings.Trim(strings.TrimSpace(r.URL.Path), "/")
		if trimmedPath != "" {
			pathParts := strings.Split(trimmedPath, "/")
			rawProductId = pathParts[len(pathParts)-1]
		}
	}

	if rawProductId == "" {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error": "invalid product id - missing path parameter",
		})
		return
	}

	decodedProductId, err := url.PathUnescape(rawProductId)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":    "invalid product id - malformed path value",
			"received": rawProductId,
		})
		return
	}
	decodedProductId = strings.TrimSpace(decodedProductId)
	decodedProductId = strings.Trim(decodedProductId, "{}\"'")

	productId, err := uuid.Parse(decodedProductId)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error":    "invalid product id - must be a valid uuid",
			"received": decodedProductId,
		})
		return
	}
	product, err := api.ProductService.GetProductById(r.Context(), productId)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			jsonutils.EncodeJson(w, r, http.StatusNotFound, map[string]any{
				"error": "product not found",
			})
			return
		}
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "unexpected error, try again later",
		})
		return
	}
	userId, ok := api.Sessions.Get(r.Context(), "AuthenticatedUserId").(uuid.UUID)
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "unexpected error, try again later",
		})
		return
	}
	room, ok := api.ensureAuctionRoom(productId, product.AuctionEnd)

	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"message": "the auction has ended",
		})
		return
	}

	conn, err := api.WsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "failed to upgrade to websocket, try again later",
		})
		return
	}

	client := services.NewClient(room, conn, userId)
	select {
	case <-room.Context.Done():
		_ = conn.Close()
		return
	case room.Register <- client:
	}

	go client.ReadEventLoop()
	go client.WriteEventLoop()
}
