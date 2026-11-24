package main

import (
	"adaptivesystemscourse/internal/config"
	"adaptivesystemscourse/internal/shared"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

func main() {
	wg := sync.WaitGroup{}

	// load config
	conf := shared.LoadConfig("subscriber")

	// start publishers
	for i := 0; i < conf.Experiment.NumberPublishers; i++ {
		wg.Add(1)
		go StartPublisher(i, *conf, &wg)
	}

	time.Sleep(5 * time.Second)

	fmt.Printf("Pulsar producers: %d  Total rate: %v msg/s Topic: %v",
		conf.Experiment.NumberPublishers,
		float64(conf.Experiment.NumberPublishers)*1.0/float64(conf.Experiment.PublishInterval)*1000.0,
		conf.Pulsar.TopicName)
	wg.Wait()
}

func StartPublisher(i int, conf config.Config, wg *sync.WaitGroup) {

	// seed generation
	rng := rand.New(rand.NewSource(42))

	// create client
	client, err := pulsar.NewClient(pulsar.ClientOptions{URL: "pulsar://" + conf.Pulsar.Host + ":" + strconv.Itoa(conf.Pulsar.Port)})
	if err != nil {
		fmt.Printf("create client: %v\n", err)
		os.Exit(0)
	}
	//defer client.Close()

	producer, err := client.CreateProducer(pulsar.ProducerOptions{Topic: conf.Pulsar.TopicName})
	if err != nil {
		log.Fatalf("create pub: %v", err)
	}
	//defer pub.Close()

	//fmt.Printf("Producer %v created at %v msg/s\n", i, 1000/mean)
	ctx := context.Background()
	for {

		payload := fmt.Sprintf("fixed-message-payload")
		//messageID, err := pub.Send(ctx, &pulsar.ProducerMessage{
		_, err := producer.Send(ctx, &pulsar.ProducerMessage{
			Payload: []byte(payload),
			//EventTime: time.Now(),
			Properties: map[string]string{
				"sent-at-ns": time.Now().UTC().Format(time.RFC3339Nano),
			},
		})
		if err != nil {
			log.Fatalf("send %q: %v", payload, err)
		}

		// configure sleep time
		t := time.Duration(shared.SampleNormal(rng, float64(conf.Experiment.PublishInterval), float64(conf.Experiment.PublishIntervalSTD)))
		time.Sleep(t * time.Millisecond)
	}
}
