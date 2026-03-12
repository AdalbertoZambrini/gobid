package api

import (
	"gobid/internal/services"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

type Api struct {
	Router         *chi.Mux
	UserService    services.UsersService
	ProductService services.ProductService
	BidsService    services.BidsService
	Sessions       *scs.SessionManager
	WsUpgrader     websocket.Upgrader
	AuctionLobby   *services.AuctionLobby
}
