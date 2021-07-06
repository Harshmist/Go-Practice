package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Item struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type itemHandlers struct {
	sync.Mutex
	store map[string]Item
}

func (h *itemHandlers) items(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.get(w, r)
		return
	case "POST":
		h.post(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed"))
		return
	}

}
func (h *itemHandlers) get(w http.ResponseWriter, r *http.Request) {
	items := make([]Item, len(h.store))

	h.Lock()
	i := 0
	for _, item := range h.store {
		items[i] = item
		i++
	}
	h.Unlock()

	jsonBytes, err := json.Marshal(items)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}
func (h *itemHandlers) getItem(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	h.Lock()
	item, ok := h.store[parts[2]]
	h.Unlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	jsonBytes, err := json.Marshal(item)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}
func (h *itemHandlers) post(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ct := r.Header.Get("content-type")
	if ct != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(fmt.Sprintf("need content-type 'application/json', but got '%s'", ct)))
		return
	}
	var item Item
	err = json.Unmarshal(bodyBytes, &item)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
	item.Key = fmt.Sprintf("%d", time.Now().UnixNano())

	h.Lock()
	h.store[item.Key] = item
	defer h.Unlock()
}

func newItemHandlers() *itemHandlers {
	return &itemHandlers{
		store: map[string]Item{},
	}
}

func main() {
	itemHandlers := newItemHandlers()
	http.HandleFunc("/items", itemHandlers.items)
	http.HandleFunc("/items/", itemHandlers.getItem)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		panic(err)
	}

}
