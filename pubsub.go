package main

import (
	"errors"
	"sync"
	"time"
)

type Topic string
type Publisher chan string
type Subscriber chan string

type PubSub struct {
	sync.Mutex
	Publishers    map[Topic]Publisher
	Subscriptions map[Topic]map[Subscriber]bool
}

func NewPubSub() *PubSub {
	return &PubSub{
		Publishers:    map[Topic]Publisher{},
		Subscriptions: map[Topic]map[Subscriber]bool{},
	}
}

func (p *PubSub) CreateTopicIfNotExists(topic Topic, createPublisher func() (Publisher, error)) error {
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
	p.Subscriptions[topic] = map[Subscriber]bool{}
	return nil
}

func (p *PubSub) removeTopic(topic Topic) {
	for subscriber := range p.Subscriptions[topic] {
		close(subscriber)
	}

	delete(p.Subscriptions, topic)
	delete(p.Publishers, topic)
}

func (p *PubSub) removeTopicIfNoSubs(topic Topic) {
	if subscribers, ok := p.Subscriptions[topic]; ok && len(subscribers) == 0 {
		p.removeTopic(topic)
	}
}

func (p *PubSub) Subscribe(sub Subscriber, topic Topic) error {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.Subscriptions[topic]; !ok {
		return errors.New("no such topic!")
	}

	p.Subscriptions[topic][sub] = true
	return nil
}

func (p *PubSub) Unsubscribe(sub Subscriber, topic Topic) error {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.Subscriptions[topic]; !ok {
		return errors.New("no such publisher!")
	}

	delete(p.Subscriptions[topic], sub)
	p.removeTopicIfNoSubs(topic)

	return nil
}

func (p *PubSub) Loop() {
	for {
		p.Lock()
		for topic, subscribers := range p.Subscriptions {
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
