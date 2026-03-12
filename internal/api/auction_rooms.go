package api

import (
	"context"
	"time"

	"gobid/internal/services"

	"github.com/google/uuid"
)

func (api *Api) ensureAuctionRoom(productID uuid.UUID, auctionEnd time.Time) (*services.AuctionRoom, bool) {
	if !auctionEnd.After(time.Now()) {
		api.AuctionLobby.Lock()
		delete(api.AuctionLobby.Rooms, productID)
		api.AuctionLobby.Unlock()
		return nil, false
	}

	api.AuctionLobby.Lock()
	room, ok := api.AuctionLobby.Rooms[productID]
	if ok {
		api.AuctionLobby.Unlock()
		return room, true
	}

	ctx, _ := context.WithDeadline(context.Background(), auctionEnd)
	room = services.NewAuctionRoom(ctx, productID, api.BidsService)
	api.AuctionLobby.Rooms[productID] = room
	api.AuctionLobby.Unlock()

	go func(createdRoom *services.AuctionRoom) {
		createdRoom.Run()
		api.AuctionLobby.Lock()
		if currentRoom, exists := api.AuctionLobby.Rooms[productID]; exists && currentRoom == createdRoom {
			delete(api.AuctionLobby.Rooms, productID)
		}
		api.AuctionLobby.Unlock()
	}(room)

	return room, true
}
