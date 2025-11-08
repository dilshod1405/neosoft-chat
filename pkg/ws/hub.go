package ws

import "sync"

type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[*Client]struct{} // conversationID(hex) -> clients
}

func NewHub() *Hub { return &Hub{rooms: make(map[string]map[*Client]struct{})} }

func (h *Hub) Join(conv string, c *Client) {
	h.mu.Lock(); defer h.mu.Unlock()
	if _, ok := h.rooms[conv]; !ok {
		h.rooms[conv] = make(map[*Client]struct{})
	}
	h.rooms[conv][c] = struct{}{}
}

func (h *Hub) Leave(conv string, c *Client) {
	h.mu.Lock(); defer h.mu.Unlock()
	if room, ok := h.rooms[conv]; ok {
		delete(room, c)
		if len(room) == 0 { delete(h.rooms, conv) }
	}
}

func (h *Hub) Broadcast(conv string, payload []byte, except *Client) {
	h.mu.RLock(); defer h.mu.RUnlock()
	if room, ok := h.rooms[conv]; ok {
		for cl := range room {
			if cl == except { continue }
			select { case cl.send <- payload: default: }
		}
	}
}
