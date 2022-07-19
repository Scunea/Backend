package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var websockets map[string]map[string]map[string]*websocket.Conn

func HandleWebsocket(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	var token = r.URL.Query().Get("token")
	var schoolId = r.URL.Query().Get("school")
	if len(token) > 0 && len(schoolId) > 0 && VerifyToken(token, db) && VerifySchool(token, schoolId, db) {
		var user, ok = GetUser(db, strings.ReplaceAll(token, "Bearer ", ""), "token")
		if ok {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Println(err)
				return
			}
			if websockets == nil {
				websockets = make(map[string]map[string]map[string]*websocket.Conn)
			}
			if websockets[schoolId] == nil {
				websockets[schoolId] = make(map[string]map[string]*websocket.Conn)
			}
			if websockets[schoolId][user.Id] == nil {
				websockets[schoolId][user.Id] = make(map[string]*websocket.Conn)
			}

			var uuid = uuid.NewString()

			conn.SetReadDeadline(time.Now().Add(time.Second * 10))

			conn.WriteMessage(websocket.PingMessage, nil)

			conn.SetPongHandler(func(string) error {
				conn.WriteMessage(websocket.PingMessage, nil)
				conn.SetReadDeadline(time.Now().Add(time.Second * 10))
				return nil
			})

			websockets[schoolId][user.Id][uuid] = conn

			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					if websockets[schoolId][user.Id][uuid] != nil {
						websockets[schoolId][user.Id][uuid].Close()
						delete(websockets[schoolId][user.Id], uuid)
					}
					break
				}
			}
		} else {
			http.Error(w, "Server error.", http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "Forbidden.", http.StatusForbidden)
		return
	}
}

func SendWebsocket(schoolId string, userId string, message interface{}) {
	if (len(schoolId) > 0 || len(userId) > 0) && message != nil {
		var messageString, err = json.Marshal(message)
		if err == nil {
			if len(schoolId) > 0 {
				if len(userId) > 0 {
					if websockets[schoolId][userId] != nil {
						for _, user := range websockets[schoolId][userId] {
							user.WriteMessage(websocket.TextMessage, []byte(messageString))
						}
					}
				} else {
					for _, users := range websockets[schoolId] {
						for _, user := range users {
							user.WriteMessage(websocket.TextMessage, []byte(messageString))
						}
					}
				}
			} else {
				for _, school := range websockets {
					for _, user := range school[userId] {
						user.WriteMessage(websocket.TextMessage, []byte(messageString))
					}
				}
			}
		}
	}
}
