package main

import (
	"errors"
	"sync"
	"time"
)

type PubSub struct {
	sync.Mutex
	Publishers  map[string]<-chan string          // publishers by topic
	Stops       map[string]chan<- struct{}        // channels the publishers listen on for a stop signal; indexed by topic
	Subscribers map[string]map[chan<- string]bool // set of subscribers by topic
}

func NewPubSub() *PubSub {
	return &PubSub{
		Publishers:  map[string]<-chan string{},
		Stops:       map[string]chan<- struct{}{},
		Subscribers: map[string]map[chan<- string]bool{},
	}
}

func (p *PubSub) CreateTopicIfNotExists(topic string, createPublisher func() (<-chan string, chan<- struct{}, error)) error {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.Publishers[topic]; ok {
		return nil
	}

	pub, stop, err := createPublisher()
	if err != nil {
		return err
	}

	p.Publishers[topic] = pub
	p.Stops[topic] = stop
	p.Subscribers[topic] = map[chan<- string]bool{}
	return nil
}

// call only from exported method to ensure p is locked!
func (p *PubSub) stopPublisherIfNoSubs(topic string) {
	if subscribers, ok := p.Subscribers[topic]; ok && len(subscribers) == 0 {
		p.Stops[topic] <- struct{}{}
		close(p.Stops[topic])
	}
}

func (p *PubSub) Subscribe(sub chan<- string, topic string) error {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.Subscribers[topic]; !ok {
		return errors.New("no such topic!")
	}

	p.Subscribers[topic][sub] = true
	return nil
}

func (p *PubSub) Unsubscribe(sub chan<- string, topic string) error {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.Subscribers[topic]; !ok {
		return errors.New("no such publisher!")
	}

	delete(p.Subscribers[topic], sub)
	p.stopPublisherIfNoSubs(topic)

	return nil
}

func (p *PubSub) Loop() {
	for {
		p.Lock()
		for topic, subscribers := range p.Subscribers {
			select {
			case message, ok := <-p.Publishers[topic]:
				if !ok {
					for subscriber := range p.Subscribers[topic] {
						close(subscriber)
					}

					delete(p.Subscribers, topic)
					delete(p.Publishers, topic)

					continue
				}

				for subscriber := range subscribers {
					subscriber <- message
				}
			case <-time.After(time.Millisecond):
				// wait at most 1ms before moving on to the next publisher
			}
		}
		p.Unlock()
	}
}
