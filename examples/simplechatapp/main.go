package main

import (
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/JosiahWitt/eventbus"
)

type Message struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Body     string   `json:"body"`
	Hashtags []string `json:"hashtags"`
}

//go:embed *.html
var staticFiles embed.FS

func main() {
	bus := eventbus.New[*Message]()

	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.FS(staticFiles)))

	mux.HandleFunc("/api/send-message", func(w http.ResponseWriter, r *http.Request) {
		msg := &Message{}
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, err.Error(), 400)
			log.Printf("Invalid message payload: %v\n", err)
			return
		}

		log.Printf("Client sending message with ID: %v\n", msg.ID)
		bus.Publish(msg, msg.Hashtags...)

		w.WriteHeader(200)
	})

	mux.HandleFunc("/api/message-stream", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", 500)
			log.Println("Streaming unsupported")
			return
		}

		hashtags := strings.Split(r.URL.Query().Get("hashtags"), ",")
		log.Printf("Client listening for hashtags: %+v\n", hashtags)

		w.Header().Add("Connection", "Keep-Alive")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Add("Content-Type", "text/event-stream")
		w.WriteHeader(200)

		sub := bus.Subscribe(hashtags...)

		go func() {
			<-r.Context().Done()
			log.Println("Client closed the connection")
			sub.Unsubscribe()
		}()

		for msg := range sub.Channel() {
			msgJSON, _ := json.Marshal(msg)
			w.Write([]byte("data: " + string(msgJSON) + "\n\n"))
			flusher.Flush()
		}
	})

	log.Println("Listening on http://localhost:1234")
	if err := http.ListenAndServe(":1234", mux); err != nil {
		log.Fatalln("Unable to start server:", err)
	}
}
