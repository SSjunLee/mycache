package main

import (
	"flag"
	"fmt"
	"log"
	"mycache"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroupAndPicker(addr string, addrs []string) (*mycache.Group, http.Handler) {
	//name string, cacheBytes int64, getter Getter
	gee := mycache.NewGroup("source", 2<<10, mycache.GetterFunc(func(key string) ([]byte, error) {
		log.Println("[SlowDB] search key", key)
		v, ok := db[key]
		if !ok {
			return nil, fmt.Errorf("%s not exist", key)
		}
		return []byte(v), nil
	}))
	picker := mycache.NewHttpPool(addr)
	gee.RegisterPeers(picker)
	picker.Set(addrs...)
	return gee, picker
}

func startCacheServer(addr string, picker http.Handler) {
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], picker))
}

func startApiServer(apiAddr string, gee *mycache.Group) {
	http.Handle("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		bv, err := gee.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(bv.ByteSlice())
	}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, ipaddr := range addrMap {
		addrs = append(addrs, ipaddr)
	}

	var serverAddr string
	if api {
		serverAddr = apiAddr
	} else {
		serverAddr = addrMap[port]
	}

	gee, picker := createGroupAndPicker(serverAddr, addrs)
	if api {
		startApiServer(serverAddr, gee)
	} else {
		startCacheServer(serverAddr, picker)
	}
}
