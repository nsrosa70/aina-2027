package main

import (
	"adaptivesystemscourse/internal/config"
	"adaptivesystemscourse/internal/shared"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	wg := sync.WaitGroup{}

	// load config
	conf := shared.LoadConfig("subscriber")

	fmt.Printf("RabbitMQ producers: %d  Total rate: %v msg/s Queue %v",
		conf.Experiment.NumberPublishers,
		float64(conf.Experiment.NumberPublishers)*1.0/float64(conf.Experiment.PublishInterval)*1000.0,
		conf.RabbitMQ.QueueName)

	// start publishers
	for i := 0; i < conf.Experiment.NumberPublishers; i++ {
		wg.Add(1)
		go StartPublisher(i, *conf, &wg)
	}
	wg.Wait()
}

func StartPublisher(id int, conf config.Config, wg *sync.WaitGroup) {
	defer wg.Done()

	// 1. Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://" + conf.RabbitMQ.User + ":" + conf.RabbitMQ.Pwd + "@" + conf.RabbitMQ.Host + ":" + strconv.Itoa(conf.RabbitMQ.Port))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// 2. Open a channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	defer ch.Close()

	// 3. Declare the same queue used by the subscriber-prefetch
	q, err := ch.QueueDeclare(
		conf.RabbitMQ.QueueName,
		true, // durable (match sub version)
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	// 4. Publish messages periodically
	rng := rand.New(rand.NewSource(42))
	for i := 0; ; i++ {
		body := fmt.Sprintf("Message %d", i)

		err = ch.Publish(
			"",     // exchange (default)
			q.Name, // routing key = queue name
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			},
		)
		if err != nil {
			log.Printf("Publish error: %v", err)
			continue
		}

		//log.Printf("Sent: %s", body)
		t := time.Duration(shared.SampleNormal(rng, float64(conf.Experiment.PublishInterval), float64(conf.Experiment.PublishIntervalSTD)))
		time.Sleep(t * time.Millisecond)
	}
}
