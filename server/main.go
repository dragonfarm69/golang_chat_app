package main

import (
	"flag"
	"log"
	"net/http"
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

	hubManager.createNewHub("temp name")
	http.HandleFunc("/", serveMainHtml)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hubId := r.URL.Query().Get("hub")

		if hubId == "" {
			http.Error(w, "404 Not found", http.StatusNotFound)
			return
		}

		hub := hubManager.getHub(hubId)
		if hub == nil {
			http.Error(w, "404 Not found", http.StatusNotFound)
			return
		}

		serveWs(hub, w, r)
	})
	http.HandleFunc("/newhub", func(w http.ResponseWriter, r *http.Request) {
		hub_id := hubManager.createNewHub("temp")
		w.Write([]byte(hub_id))
	})
	http.HandleFunc("/hublist", func(w http.ResponseWriter, r *http.Request) {
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
	http.HandleFunc("/disconnect/", func(w http.ResponseWriter, r *http.Request) {
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

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("error when starting server: ", err)
	}
}
