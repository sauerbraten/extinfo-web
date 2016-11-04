package pubsub

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

type PubSub struct {
	sync.Mutex
	Publishers  map[string]<-chan []byte                   // publishers' update channel by topic
	Stops       map[string]chan<- struct{}                 // channels the publishers listen on for a stop signal; indexed by topic
	Subscribers map[string]map[<-chan Update]chan<- Update // set of subscribers by topic; subs are identified by the receiving end provided by Subscribe()

	notifyAboutTopic chan string // channel on which publishers notify PubSub about news for a certain topic
}

func NewPubSub() *PubSub {
	p := &PubSub{
		Publishers:       map[string]<-chan []byte{},
		Stops:            map[string]chan<- struct{}{},
		Subscribers:      map[string]map[<-chan Update]chan<- Update{},
		notifyAboutTopic: make(chan string),
	}

	go p.loop()

	return p
}

var p = NewPubSub()

func (p *PubSub) createTopicIfNotExists(topic string, useNewPublisher func(Publisher) error) error {
	if _, ok := p.Publishers[topic]; ok {
		return nil
	}

	publisher, updates, stop := newPublisher(topic, p.notifyAboutTopic)
	err := useNewPublisher(publisher)
	if err != nil {
		return err
	}

	p.Publishers[topic] = updates
	p.Stops[topic] = stop
	p.Subscribers[topic] = map[<-chan Update]chan<- Update{}
	return nil
}

func (p *PubSub) stopPublisherIfNoSubs(topic string) {
	if subscribers, ok := p.Subscribers[topic]; !ok || len(subscribers) != 0 {
		return
	}

	close(p.Stops[topic])
}

// assumes that the publisher closed its updates channel
func (p *PubSub) removeTopic(topic string) {
	for _, subscriber := range p.Subscribers[topic] {
		close(subscriber)
	}

	delete(p.Subscribers, topic)
	delete(p.Publishers, topic)
	delete(p.Stops, topic)
}

func (p *PubSub) Subscribe(topic string, useNewPublisher func(Publisher) error) (<-chan Update, error) {
	p.Lock()
	defer p.Unlock()

	err := p.createTopicIfNotExists(topic, useNewPublisher)
	if err != nil {
		return nil, err
	}

	updates := make(chan Update)

	p.Subscribers[topic][updates] = updates
	return updates, nil
}

func Subscribe(topic string, useNewPublisher func(Publisher) error) (<-chan Update, error) {
	return p.Subscribe(topic, useNewPublisher)
}

func (p *PubSub) Unsubscribe(sub <-chan Update, topic string) error {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.Publishers[topic]; !ok {
		return errors.New("no such topic")
	}

	delete(p.Subscribers[topic], sub)
	p.stopPublisherIfNoSubs(topic)

	return nil
}

func Unsubscribe(sub <-chan Update, topic string) error {
	return p.Unsubscribe(sub, topic)
}

func (p *PubSub) loop() {
	statusTicker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case topic := <-p.notifyAboutTopic:
			p.Lock()

			message, ok := <-p.Publishers[topic]
			if ok {
				for _, subscriber := range p.Subscribers[topic] {
					select {
					case subscriber <- Update{
						Topic:   topic,
						Content: message,
					}:
					case <-time.After(10 * time.Millisecond):
						// 10ms timeout for each subscriber to receive
						log.Println("send timed out for subscriber", subscriber)
					}
				}
			} else {
				p.removeTopic(topic)
			}

			p.Unlock()

		case <-statusTicker.C:
			if len(p.Publishers) > 0 {
				log.Println("polling", len(p.Publishers), "servers")
				log.Println("serving", len(p.Subscribers), "subscribers")
			}
		}
	}
}
