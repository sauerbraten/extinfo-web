package main

import (
	"errors"
	"log"
	"sync"
	"time"
)

type Update struct {
	Topic   string
	Content []byte
}

type Publisher struct {
	topic        string
	notifyPubSub chan<- string
	updates      chan<- []byte
	Stop         <-chan struct{}
}

func NewPublisher(topic string, notifyPubSub chan<- string) (Publisher, <-chan []byte, chan<- struct{}) {
	updates := make(chan []byte)
	stop := make(chan struct{})

	p := Publisher{
		topic:        topic,
		notifyPubSub: notifyPubSub,
		updates:      updates,
		Stop:         stop,
	}

	return p, updates, stop
}

func (p *Publisher) Publish(update []byte) {
	p.notifyPubSub <- p.topic
	p.updates <- update
}

func (p *Publisher) Close() {
	close(p.updates)
}

type PubSub struct {
	sync.Mutex
	Publishers  map[string]<-chan []byte          // publishers' update channel by topic
	Stops       map[string]chan<- struct{}        // channels the publishers listen on for a stop signal; indexed by topic
	Subscribers map[string]map[chan<- Update]bool // set of subscribers by topic

	notifyAboutTopic chan string // channel on which publishers notify PubSub about news for a certain topic
}

func NewPubSub() *PubSub {
	return &PubSub{
		Publishers:       map[string]<-chan []byte{},
		Stops:            map[string]chan<- struct{}{},
		Subscribers:      map[string]map[chan<- Update]bool{},
		notifyAboutTopic: make(chan string),
	}
}

// CreateTopicIfNotExists returns nil if there's already a publisher for the specified topic. Otherwise, it creates a publisher and passes it to useNewPublisher.
// The useNewPublisher function may return an error if using the publisher failed. That error is returned to the caller of CreateTopicIfNotExists.
func (p *PubSub) CreateTopicIfNotExists(topic string, useNewPublisher func(Publisher) error) error {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.Publishers[topic]; ok {
		return nil
	}

	publisher, updates, stop := NewPublisher(topic, pubsub.notifyAboutTopic)
	err := useNewPublisher(publisher)
	if err != nil {
		return err
	}

	p.Publishers[topic] = updates
	p.Stops[topic] = stop
	p.Subscribers[topic] = map[chan<- Update]bool{}
	return nil
}

// call only from exported method to ensure p is locked!
func (p *PubSub) removeTopicIfNoSubs(topic string) {
	if subscribers, ok := p.Subscribers[topic]; !ok || len(subscribers) != 0 {
		return
	}

	close(p.Stops[topic])
	delete(p.Subscribers, topic)
	delete(p.Publishers, topic)
	delete(p.Stops, topic)
}

func (p *PubSub) Subscribe(sub chan<- Update, topic string) error {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.Subscribers[topic]; !ok {
		return errors.New("no such topic")
	}

	p.Subscribers[topic][sub] = true
	return nil
}

func (p *PubSub) Unsubscribe(sub chan<- Update, topic string) error {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.Subscribers[topic]; !ok {
		return errors.New("no such publisher")
	}

	delete(p.Subscribers[topic], sub)
	p.removeTopicIfNoSubs(topic)

	return nil
}

func (p *PubSub) Loop() {
	statusTicker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case topic := <-p.notifyAboutTopic:
			p.Lock()

			message, ok := <-p.Publishers[topic]
			if !ok {
				for subscriber := range p.Subscribers[topic] {
					close(subscriber)
				}

				delete(p.Subscribers, topic)
				delete(p.Publishers, topic)

				continue
			}

			for subscriber := range p.Subscribers[topic] {
				select {
				case subscriber <- Update{
					Topic:   topic,
					Content: message,
				}:
				case <-time.After(10 * time.Millisecond):
				}
			}

			p.Unlock()

		case <-statusTicker.C:
			log.Println("polling", len(pubsub.Publishers), "servers")
			log.Println("serving", len(pubsub.Subscribers), "subscribers")
		}
	}
}
