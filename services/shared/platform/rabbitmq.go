package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

const EventsExchange = "newfeed.events"

type RabbitPublisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewRabbitPublisher(url string) (*RabbitPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := ch.ExchangeDeclare(EventsExchange, "topic", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}
	return &RabbitPublisher{conn: conn, ch: ch}, nil
}

func (p *RabbitPublisher) Publish(ctx context.Context, routingKey string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return p.ch.PublishWithContext(ctx, EventsExchange, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now().UTC(),
		Body:         body,
	})
}

func (p *RabbitPublisher) Close() {
	_ = p.ch.Close()
	_ = p.conn.Close()
}

type RabbitConsumer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	db   *pgxpool.Pool
}

type EventHandler func(context.Context, EventEnvelope) error

func NewRabbitConsumer(url string, db *pgxpool.Pool) (*RabbitConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := ch.ExchangeDeclare(EventsExchange, "topic", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}
	return &RabbitConsumer{conn: conn, ch: ch, db: db}, nil
}

func (c *RabbitConsumer) Consume(ctx context.Context, queue, routingKey string, handler EventHandler) error {
	dlq := queue + ".dlq"
	if _, err := c.ch.QueueDeclare(dlq, true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := c.ch.QueueDeclare(queue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": dlq,
	}); err != nil {
		return err
	}
	if err := c.ch.QueueBind(queue, routingKey, EventsExchange, false, nil); err != nil {
		return err
	}
	deliveries, err := c.ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case delivery, ok := <-deliveries:
				if !ok {
					return
				}
				if err := c.handleDelivery(ctx, delivery, handler); err != nil {
					_ = delivery.Nack(false, false)
				} else {
					_ = delivery.Ack(false)
				}
			}
		}
	}()
	return nil
}

func (c *RabbitConsumer) handleDelivery(ctx context.Context, delivery amqp.Delivery, handler EventHandler) error {
	var event EventEnvelope
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		return err
	}
	if event.EventID == "" || event.EventName == "" {
		return fmt.Errorf("invalid event envelope")
	}
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `INSERT INTO processed_events(event_id, event_name) VALUES($1,$2) ON CONFLICT DO NOTHING`, event.EventID, event.EventName)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return tx.Commit(ctx)
	}
	if err := handler(ctx, event); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (c *RabbitConsumer) Close() {
	_ = c.ch.Close()
	_ = c.conn.Close()
}
