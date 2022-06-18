package main

import (
  "net/http"
  "encoding/json"
  "sync"
  "io/ioutil"
  "fmt"
  "time"
)

type Item struct {
  Id string `json:"id"`
  Name string `json:"name"`
  Quantity int `json:"quantity"`
}

type itemHandlers struct {
  sync.Mutex
  store map[string]Item
}

func (h * itemHandlers) getItems(w http.ResponseWriter, r *http.Request) {
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
  w.Write(jsonBytes)
}

func (h * itemHandlers) createItem(w http.ResponseWriter, r *http.Request) {
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
    w.Write([]byte("Only json is supported"))
    return
  }

  var item Item
  err = json.Unmarshal(bodyBytes, &item)

  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(err.Error()))
    return
  }

  if item.Quantity == 0 || item.Name == "" {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte("Invalid Request"))
    return
  }

  item.Id = fmt.Sprintf("%d", time.Now().UnixNano())

  h.Lock()
  h.store[item.Id] = item
  defer h.Unlock()

}

func newItemHandlers() *itemHandlers {
  return &itemHandlers {
    store: map[string]Item{
      "id1": Item {
        Id: "id1",
        Name: "Carrots",
        Quantity: 10,
      },
    },
  }
}


func main() {
  itemHandlers := newItemHandlers()

  http.HandleFunc("/items/list", itemHandlers.getItems)
  http.HandleFunc("/items/create", itemHandlers.createItem)
  err := http.ListenAndServe(":8081", nil)
  if err != nil {
    panic(err)
  }
}
