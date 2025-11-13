package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/rs/cors"
)

var addr = flag.String("addr", ":8080", "chat server service")

func serveMainHtml(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "404 Not Found", http.StatusNotFound)
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "index.html")
}

func main() {
	flag.Parse()
	hubManager := newHubManager()
	// hub := newHub()
	// go hub.run()

	mux := http.NewServeMux()

	hubManager.createNewHub("temp name")
	mux.HandleFunc("/", serveMainHtml)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hubId := r.URL.Query().Get("hub")
		client_name := r.URL.Query().Get("username")
		if client_name == "" {
			http.Error(w, "Unknown User", http.StatusBadRequest)
			return
		}

		if hubId == "" {
			http.Error(w, "404 Not found", http.StatusNotFound)
			return
		}

		hub := hubManager.getHub(hubId)
		if hub == nil {
			http.Error(w, "404 Not found", http.StatusNotFound)
			return
		}

		serveWs(hubManager, w, r, client_name, hub)
	})
	mux.HandleFunc("/newhub", func(w http.ResponseWriter, r *http.Request) {
		hub_id := hubManager.createNewHub("temp")
		w.Write([]byte(hub_id))
	})
	mux.HandleFunc("/hublist", func(w http.ResponseWriter, r *http.Request) {
		lists := hubManager.getHubListIds()
		s := ""
		for i, id := range lists {
			if i > 0 {
				s += ","
			}
			s += id
		}
		w.Write([]byte(s))
	})
	mux.HandleFunc("/disconnect/", func(w http.ResponseWriter, r *http.Request) {
		log.Print("diconnecting")
		hubId := r.URL.Query().Get("hub")
		clientId := r.URL.Query().Get("client")
		log.Println(hubId)
		log.Println(clientId)

		if hubId == "" || clientId == "" {
			http.Error(w, "Can't be empty", http.StatusNotFound)
			return
		}

		hub := hubManager.getHub(hubId)
		if hub == nil {
			http.Error(w, "404 Not found", http.StatusNotFound)
			return
		}

		hub.disconnectClient(clientId)
	})
	mux.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) {
		hubId := r.URL.Query().Get("hub")
		clientId := r.URL.Query().Get("client")

		if hubId == "" || clientId == "" {
			http.Error(w, "Can't be empty", http.StatusNotFound)
			return
		}

		hub := hubManager.getHub(hubId)
		if hub == nil {
			http.Error(w, "404 Not found", http.StatusNotFound)
			return
		}

		// hub.disconnectClient(clientId)
		// serveWs(hub, w, r)
		w.WriteHeader(http.StatusOK)
	})

	//Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Allow all origins
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})
	handler := c.Handler(mux)

	err := http.ListenAndServe(*addr, handler)
	if err != nil {
		log.Fatal("error when starting server: ", err)
	}
}
