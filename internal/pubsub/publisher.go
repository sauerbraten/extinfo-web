package pubsub

type Publisher struct {
	topic        string
	notifyPubSub chan<- string
	updates      chan<- []byte
	Stop         <-chan struct{}
}

func newPublisher(topic string, notifyPubSub chan<- string) (Publisher, <-chan []byte, chan<- struct{}) {
	updates := make(chan []byte, 1)
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
	p.notifyPubSub <- p.topic
}
