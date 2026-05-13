package database

import (
	"encoding/json"
	"fmt"

	"DeepSight/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

var rabbitMQ *RabbitMQ

type RabbitMQ struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	Exchange  string
	QueueName string
	Key       string
	URL       string
}

func InitializeRabbitMQ(cfg *config.RabbitMQConfig) error {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
	)

	conn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to connect rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("failed to open rabbitmq channel: %w", err)
	}

	if cfg.Exchange != "" {
		if err := ch.ExchangeDeclare(
			cfg.Exchange,
			"direct",
			true,
			false,
			false,
			false,
			nil,
		); err != nil {
			_ = ch.Close()
			_ = conn.Close()
			return fmt.Errorf("failed to declare rabbitmq exchange %q: %w", cfg.Exchange, err)
		}
	}

	rabbitMQ = &RabbitMQ{
		conn:      conn,
		channel:   ch,
		Exchange:  cfg.Exchange,
		QueueName: cfg.Queue,
		Key:       cfg.BindingKey,
		URL:       url,
	}

	return nil
}

func GetRabbitMQ() *RabbitMQ {
	return rabbitMQ
}

func CloseRabbitMQ() error {
	if rabbitMQ == nil {
		return nil
	}

	var firstErr error
	if rabbitMQ.channel != nil {
		if err := rabbitMQ.channel.Close(); err != nil {
			firstErr = err
		}
	}
	if rabbitMQ.conn != nil {
		if err := rabbitMQ.conn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	rabbitMQ = nil
	return firstErr
}

// PublishJSON 发布 JSON 格式的消息
func (r *RabbitMQ) PublishJSON(data interface{}) error {
	if r == nil || r.channel == nil {
		return fmt.Errorf("rabbitmq is not initialized")
	}

	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if r.Exchange != "" {
		if err := r.channel.ExchangeDeclare(
			r.Exchange,
			"direct",
			true,
			false,
			false,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("failed to declare rabbitmq exchange %q: %w", r.Exchange, err)
		}
	}

	if err := r.channel.Publish(
		r.Exchange,
		r.Key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	); err != nil {
		return fmt.Errorf("failed to publish rabbitmq message: %w", err)
	}

	return nil
}

func (r *RabbitMQ) ReceiveRouting() (<-chan amqp.Delivery, error) {
	if r == nil || r.channel == nil {
		return nil, fmt.Errorf("rabbitmq is not initialized")
	}

	if err := r.channel.ExchangeDeclare(
		r.Exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return nil, fmt.Errorf("failed to declare rabbitmq exchange %q: %w", r.Exchange, err)
	}

	q, err := r.channel.QueueDeclare(
		r.QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare rabbitmq queue: %w", err)
	}

	if err := r.channel.QueueBind(
		q.Name,
		r.Key,
		r.Exchange,
		false,
		nil,
	); err != nil {
		return nil, fmt.Errorf("failed to bind rabbitmq queue: %w", err)
	}

	msgs, err := r.channel.Consume(
		q.Name,
		"",
		false, // 改为手动确认，确保处理成功后才 ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume rabbitmq queue: %w", err)
	}

	return msgs, nil
}
