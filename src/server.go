package main

import (
  "net/http"
  "encoding/json"
  "sync"
  "io/ioutil"
  "fmt"
  "time"
  "strings"
  "os"
  "math/rand"
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

func (h * itemHandlers) getRandomItem(w http.ResponseWriter, r *http.Request) {
  ids := make([]string, len(h.store))
  h.Lock()

  i := 0
  for id := range h.store {
    ids[i] = id
    i++
  }

  defer h.Unlock()

  var target string
  if len(ids) == 0 {
    w.WriteHeader(http.StatusNotFound)
    
  } else if len(ids) == 1{
    target = ids[0]
    
  } else {

    rand.Seed(time.Now().UnixNano())
  
    target = ids[rand.Intn(len(ids))]
  
  }
  
  w.Header().Add("location", fmt.Sprintf("/items/%s", target))
  w.WriteHeader(http.StatusFound)

}

func (h * itemHandlers) getItem(w http.ResponseWriter, r *http.Request) {
  parts := strings.Split(r.URL.String(), "/")

  if len(parts) != 3 {
    w.WriteHeader(http.StatusNotFound)
    return
  }

  if parts[2] == "random" {
    h.getRandomItem(w,r)
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
      "1655570749194813500": Item {
        Id: "1655570749194813500",
        Name: "Carrots",
        Quantity: 10,
      },
    },
  }
}

type adminPortal struct {
  password string
}

func newAdminPortal() *adminPortal{
  password := os.Getenv("ADMIN_PASSWORD")

  if password == "" {
    panic("Admin password not set")
  }

  return &adminPortal{password: password}
}

func (a adminPortal) handler(w http.ResponseWriter, r *http.Request) {
  user, pass, ok := r.BasicAuth()

  if !ok || user != "admin" || pass != a.password {
    w.WriteHeader(http.StatusUnauthorized)
    w.Write([]byte("Not authorized"))
    return
  }

  w.Write([]byte("Super secret admin portal"))
}

func main() {
  admin := newAdminPortal()
  itemHandlers := newItemHandlers()

  http.HandleFunc("/items/list", itemHandlers.getItems)
  http.HandleFunc("/items/", itemHandlers.getItem)
  http.HandleFunc("/items/create", itemHandlers.createItem)

  http.HandleFunc("/admin", admin.handler)

  err := http.ListenAndServe(":8081", nil)
  if err != nil {
    panic(err)
  }
}
