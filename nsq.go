package next

import (
	"fmt"
	"github.com/nsqio/go-nsq"
)

type Nsq struct {
	addr string
}

type NsqMessage struct {
	*nsq.Message
}

func NewNsq() *Nsq {
	return &Nsq{}
}

func (n *Nsq) Open(addr string) {
	n.addr = addr
}

func (n *Nsq) Publish(topicName string, message []byte) error {
	// Create the configuration object and set the maxInFlight
	cfg := nsq.NewConfig()
	cfg.MaxInFlight = 8

	// Create the producer
	p, err := nsq.NewProducer(n.addr, cfg)
	if err != nil {
		return err
	}
	return p.Publish(topicName, message)
}

func (n *Nsq) Subscribe(topicName, channelName string, hander nsq.HandlerFunc) error {
	fmt.Printf("Subscribe on %s/%s\n", topicName, channelName)

	// Create the configuration object and set the maxInFlight
	cfg := nsq.NewConfig()
	cfg.MaxInFlight = 8

	// Create the consumer with the given topic and chanel names
	r, err := nsq.NewConsumer(topicName, channelName, cfg)
	if err != nil {
		return err
	}

	// Set the handler
	r.AddHandler(hander)

	// Connect to the NSQ daemon
	if err := r.ConnectToNSQD(n.addr); err != nil {
		return err
	}

	// Wait for the consumer to stop.
	<-r.StopChan
	return nil
}
