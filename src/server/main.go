package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type counters struct {
	sync.Mutex
	view  int
	click int
	selection string
}

type values struct {
	view int
	click int
}

var (
	c = counters{}
	content = []string{"sports", "entertainment", "business", "education"}
)

var counterQueue chan counters
var mockStore map[string]values

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to EQ Works 😎")
}

func getSelection(data string) string {
	t := time.Now()
	key := fmt.Sprintf("%s:%d-%02d-%02d %02d:%02d", data, t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
	return key
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	data := content[rand.Intn(len(content))]

	c.Lock()
	c.view++
	c.selection = getSelection(data)
	c.Unlock()

	err := processRequest(r)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(400)
		return
	}

	// simulate random click call
	if rand.Intn(100) < 50 {
		processClick(data)
	}

	c.Lock()
	counterQueue <- c
	c.Unlock()
}

func processRequest(r *http.Request) error {
	time.Sleep(time.Duration(rand.Int31n(50)) * time.Millisecond)
	return nil
}

func processClick(data string) error {
	c.Lock()
	c.click++
	c.Unlock()

	return nil
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	if !isAllowed() {
		w.WriteHeader(429)
		return
	}
}

func isAllowed() bool {

	return true
}

func uploadCounters() error {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <- ticker.C:
			cData := <- counterQueue
			mockStore[cData.selection] = values{view: cData.view, click: cData.click}
		}
	}
	return nil
}

func main() {
	http.HandleFunc("/", welcomeHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/stats/", statsHandler)

	counterQueue = make(chan counters, 5)
	mockStore = make(map[string]values)
	go uploadCounters()

	log.Fatal(http.ListenAndServe(":8080", nil))
}
