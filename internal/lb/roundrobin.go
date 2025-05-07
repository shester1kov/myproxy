package lb

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

type BackendStatus struct {
	URL   string
	Alive bool
	mutex sync.RWMutex
}

type RoundRobin struct {
	backends []*BackendStatus
	current  int
	mu       sync.Mutex
}

func NewRoundRobin(backends []string) *RoundRobin {
	rr := &RoundRobin{
		backends: make([]*BackendStatus, 0, len(backends)),
		current:  0,
	}

	for _, b := range backends {
		rr.backends = append(
			rr.backends,
			&BackendStatus{
				URL:   b,
				Alive: true,
			},
		)
	}

	go rr.HealthCheck()
	return rr
}

func (r *RoundRobin) HealthCheck() {
	ticker := time.NewTicker(10 * time.Second)
	client := http.Client{Timeout: 2 * time.Second}

	for range ticker.C {
		for _, backend := range r.backends {
			go func(b *BackendStatus) {
				resp, err := client.Get(b.URL + "/health")
				b.mutex.Lock()
				defer b.mutex.Unlock()

				if err != nil || resp.StatusCode != http.StatusOK {
					b.Alive = false
					//log.Printf("Сервер %s не работает", b.URL)
				} else {
					b.Alive = true
					//log.Printf("Сервер %s работает", b.URL)
				}
			}(backend)
		}
	}
}

func (r *RoundRobin) GetNext() (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.backends) == 0 {
		return "", errors.New("нет доступных серверов")
	}

	start := r.current

	for {
		backend := r.backends[r.current]
		backend.mutex.RLock()
		alive := backend.Alive
		backend.mutex.RUnlock()

		r.current = (r.current + 1) % len(r.backends)

		if alive {
			return backend.URL, nil
		}

		if r.current == start {
			break
		}
	}

	return "", errors.New("ни один сервер не доступен")

}
