package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
	"encoding/json"
	"strconv"
)

type counters struct {
	sync.Mutex
	view  int
	click int
	selection string
}

type values struct {
	View string `json:"view"`
	Click string `json:"click"`
}

type ipTracker struct {
	sync.Mutex
	ipAddress string
	startTime time.Time
	count int
}

const requestInterval = 60
const requestLimit = 10

var (
	c = counters{}
	content = []string{"sports", "entertainment", "business", "education"}
)

var storeMutex = &sync.Mutex{}
var counterQueue chan counters
var mockStore map[string]values
var requestStore map[string]ipTracker

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	if !isAllowed(r.RemoteAddr) {
		http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
		return
	}
	fmt.Fprint(w, "Welcome to EQ Works 😎")
}

func getSelection(data string) string {
	t := time.Now()
	key := fmt.Sprintf("%s:%d-%02d-%02d %02d:%02d", data, t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
	return key
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	if !isAllowed(r.RemoteAddr) {
		http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
		return
	}
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

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(204)
	return
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

// Returns statistics
func statsHandler(w http.ResponseWriter, r *http.Request) {
	if !isAllowed(r.RemoteAddr) {
		http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(mockStore)
	return
}

// Rate Limitar
func isAllowed(ipAddress string) bool {
	if val, ok := requestStore[ipAddress]; ok {
		currentTime := time.Now()
		if val.count < requestLimit {
			val.count++
			requestStore[ipAddress] = val
			return true
		} else {
			timeDelta := currentTime.Sub(val.startTime).Seconds()
			if timeDelta <= requestInterval {
				return false
			} else {
				val.count = 1
				val.startTime = currentTime
				requestStore[ipAddress] = val
				return true
			}
		}
	} else {
		value := ipTracker{ipAddress: ipAddress, startTime: time.Now(), count: 1}
		requestStore[ipAddress] = value
	}
	return true
}

// Function Uploading Counters
func uploadCounters() error {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <- ticker.C:
			cData := <- counterQueue
			storeMutex.Lock()
			mockStore[cData.selection] = values{View: strconv.Itoa(cData.view), Click: strconv.Itoa(cData.click)}
			storeMutex.Unlock()
		}
	}
	return nil
}

func main() {
	http.HandleFunc("/", welcomeHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/stats/", statsHandler)

	counterQueue = make(chan counters, 20)
	mockStore = make(map[string]values)
	requestStore = make(map[string]ipTracker)
	go uploadCounters()

	log.Fatal(http.ListenAndServe(":8080", nil))
}
