package ws

import "sync"

type Hub struct {
	mu sync.RWMutex
	rooms map[string]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		rooms: map[string]map[*Client]struct{}{},
	}
}

func (h *Hub) Join(room string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[room]; !ok {
		h.rooms[room] = map[*Client]struct{}{}
	}
	h.rooms[room][c] = struct{}{}
}

func (h *Hub) Leave(room string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if r, ok := h.rooms[room]; ok {
		delete(r, c)
		if len(r) == 0 {
			delete(h.rooms, room)
		}
	}
}

func (h *Hub) Broadcast(room string, msg []byte, except *Client) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if r, ok := h.rooms[room]; ok {
		for cl := range r {
			if cl == except { continue }
			select { case cl.send <- msg: default: }
		}
	}
}
