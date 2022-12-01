package rabbit

import (
	"context"
	"encoding/json"

	"github.com/streadway/amqp"
)

type Rabbit struct {
	chanName string
	channel  *amqp.Channel
	queue    amqp.Queue
}

func New(channelName string) (*Rabbit, error) {
	conn, err := amqp.Dial("amqp://user:bitnami@localhost:5672")
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	queue, err := channel.QueueDeclare(
		channelName,
		true,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &Rabbit{
		chanName: channelName,
		channel:  channel,
		queue:    queue,
	}, nil
}

func (r *Rabbit) Publish(ctx context.Context, msg any) error {
	raw, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if err = r.channel.Publish("", r.chanName, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         raw,
	}); err != nil {
		return err
	}

	return nil
}
