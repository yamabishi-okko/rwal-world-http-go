package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	// デモ用: どのオリジンからでも許可（本番は制限してね）
	CheckOrigin: func(r *http.Request) bool { return true },
}

type client struct {
	conn *websocket.Conn
	room string
}

type hub struct {
	mu      sync.Mutex
	rooms   map[string]map[*client]struct{} // room -> set of clients
}

func newHub() *hub {
	return &hub{rooms: make(map[string]map[*client]struct{})}
}

func (h *hub) join(room string, c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*client]struct{})
	}
	h.rooms[room][c] = struct{}{}
}

func (h *hub) leave(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	set := h.rooms[c.room]
	if set == nil {
		return
	}
	delete(set, c)
	if len(set) == 0 {
		delete(h.rooms, c.room)
	}
}

func (h *hub) broadcastExceptSender(c *client, msg []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for other := range h.rooms[c.room] {
		if other == c {
			continue
		}
		_ = other.conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func main() {
	h := newHub()

	// シグナリング用WS: /ws?room=lobby など
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		room := r.URL.Query().Get("room")
		if room == "" {
			room = "lobby"
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		c := &client{conn: conn, room: room}
		h.join(room, c)
		log.Printf("joined room=%s", room)

		defer func() {
			h.leave(c)
			conn.Close()
			log.Printf("left room=%s", room)
		}()

		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			// 受け取ったJSON（offer/answer/candidate）を同じ部屋の他クライアントへ転送
			h.broadcastExceptSender(c, data)
		}
	})

	// 静的ファイル
	http.Handle("/", http.FileServer(http.Dir("./public")))

	addr := ":18891"
	log.Printf("WebRTC(DataChannel) demo => http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
