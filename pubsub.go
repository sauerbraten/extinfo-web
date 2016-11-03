package main

import (
	"errors"
	"sync"
	"time"
)

type PubSub struct {
	sync.Mutex
	Publishers  map[string]chan string
	Subscribers map[string]map[chan string]bool
}

func NewPubSub() *PubSub {
	return &PubSub{
		Publishers:  map[string]chan string{},
		Subscribers: map[string]map[chan string]bool{},
	}
}

func (p *PubSub) CreateTopicIfNotExists(topic string, createPublisher func() (chan string, error)) error {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.Publishers[topic]; ok {
		return nil
	}

	pub, err := createPublisher()
	if err != nil {
		return err
	}

	p.Publishers[topic] = pub
	p.Subscribers[topic] = map[chan string]bool{}
	return nil
}

func (p *PubSub) removeTopic(topic string) {
	for subscriber := range p.Subscribers[topic] {
		close(subscriber)
	}

	delete(p.Subscribers, topic)
	delete(p.Publishers, topic)
}

func (p *PubSub) removeTopicIfNoSubs(topic string) {
	if subscribers, ok := p.Subscribers[topic]; ok && len(subscribers) == 0 {
		p.removeTopic(topic)
	}
}

func (p *PubSub) Subscribe(sub chan string, topic string) error {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.Subscribers[topic]; !ok {
		return errors.New("no such topic!")
	}

	p.Subscribers[topic][sub] = true
	return nil
}

func (p *PubSub) Unsubscribe(sub chan string, topic string) error {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.Subscribers[topic]; !ok {
		return errors.New("no such publisher!")
	}

	delete(p.Subscribers[topic], sub)
	p.removeTopicIfNoSubs(topic)

	return nil
}

func (p *PubSub) Loop() {
	for {
		p.Lock()
		for topic, subscribers := range p.Subscribers {
			select {
			case message, ok := <-p.Publishers[topic]:
				if !ok {
					p.removeTopic(topic)
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
