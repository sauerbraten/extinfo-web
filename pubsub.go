package main

import (
	"errors"
	"sync"
)

type Update struct {
	Topic   string
	Content []byte
}

type Publisher struct {
	Topic        string
	NotifyPubSub chan<- string
	Updates      chan<- []byte
	Stop         <-chan struct{}
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
		notifyAboutTopic: make(chan string, 10),
	}
}

// CreateTopicIfNotExists returns nil if there's already a publisher for the specified topic. Otherwise, it creates a publisher using the createPublisher function,
// passing it the topic and a notifications channel. Everytime the publisher wants to send data about a topic, he has to get the attention of PubSub's main loop by
// sending the topic into the notificatino channel.
// The createPublisher function has to return the receiving end of a channel on which it sends its updates, the sending end of a channel on which it listens for a
// stop signal, and an error if creating the publisher failed. That error is returned to the caller of CreateTopicIfNotExists.
func (p *PubSub) CreateTopicIfNotExists(topic string, createPublisher func(topic string, notify chan<- string) (<-chan []byte, chan<- struct{}, error)) error {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.Publishers[topic]; ok {
		return nil
	}

	pub, stop, err := createPublisher(topic, p.notifyAboutTopic)
	if err != nil {
		return err
	}

	p.Publishers[topic] = pub
	p.Stops[topic] = stop
	p.Subscribers[topic] = map[chan<- Update]bool{}
	return nil
}

// call only from exported method to ensure p is locked!
func (p *PubSub) stopPublisherIfNoSubs(topic string) {
	if subscribers, ok := p.Subscribers[topic]; ok && len(subscribers) == 0 {
		p.Stops[topic] <- struct{}{}
		close(p.Stops[topic])
	}
}

func (p *PubSub) Subscribe(sub chan<- Update, topic string) error {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.Subscribers[topic]; !ok {
		return errors.New("no such topic!")
	}

	p.Subscribers[topic][sub] = true
	return nil
}

func (p *PubSub) Unsubscribe(sub chan<- Update, topic string) error {
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
		topic := <-p.notifyAboutTopic

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
			subscriber <- Update{
				Topic:   topic,
				Content: message,
			}
		}

		p.Unlock()
	}
}
