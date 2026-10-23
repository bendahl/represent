package main

import "sync"

type Broadcaster struct {
	mu      sync.Mutex
	clients map[chan int]struct{}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[chan int]struct{}),
	}
}

func (b *Broadcaster) Subscribe() chan int {
	ch := make(chan int, 1)

	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()

	return ch
}

func (b *Broadcaster) Unsubscribe(ch chan int) {
	b.mu.Lock()
	delete(b.clients, ch)
	b.mu.Unlock()
}

func (b *Broadcaster) Publish(slideNum int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for ch := range b.clients {
		// Discard old event
		select {
		case <-ch:
		default:
		}

		// Always publish latest slide number
		select {
		case ch <- slideNum:
		default:
		}
	}
}
