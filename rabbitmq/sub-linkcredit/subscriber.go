package main

import (
	"adaptivesystemscourse/internal/config"
	"adaptivesystemscourse/internal/control/mico"
	"adaptivesystemscourse/internal/shared"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	amqp "github.com/Azure/go-amqp"
)

func main() {

	// load config
	conf := shared.LoadConfig("subscriber")

	// configure some variables
	conf.Experiment.CurrentReceiver = conf.Experiment.InitialLinkedCredit
	//conf.PIDHysteresis.HysteresisBand = conf.Experiment.Setpoint * conf.PIDHysteresis.HysteresisBand

	switch conf.Experiment.ExecutionType {
	case shared.ClosedLoop:
		RunClosedLoop(*conf)
	case shared.MultiClosedLoop:
		RunClosedLoop(*conf)
	case shared.OpenLoop:
		RunOpenLoop(*conf)
	default:
		fmt.Println("execution type unknown")
		os.Exit(1)
	}
	fmt.Scanln()
}

func RunOpenLoop(conf config.Config) {

	// create context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// create receivers
	receivers, err := CreateReceivers(conf, ctx)
	if err != nil {
		log.Fatalf("Create receivers: %v", err)
		os.Exit(0)
	}

	// configure MICO
	m := mico.MICO{}
	err = m.Initialise(conf)
	if err != nil {
		fmt.Printf("Initialise mico: %v\n", err)
	}

	// initialisations
	ticker := time.NewTicker(time.Duration(conf.Experiment.MonitorInterval) * time.Second)
	nMsgs := 0
	nSamples := 0
	u := conf.Experiment.InitialLinkedCredit
	currentReceiver := u - 1

	// loop
	fmt.Printf("************* Begin of Experiment [%v] ************* \n", conf.Experiment.ControllerType)
	for {
		select {
		case <-ticker.C:
			// compute receiveiving rate
			r := float64(nMsgs) / float64(time.Duration(conf.Experiment.MonitorInterval)*time.Second)

			fmt.Printf("%v;%v;%v\n", time.Now(), int32(u), r)

			if nSamples >= conf.Experiment.SampleSize-1 {
				nSamples = 0
				u++
				currentReceiver = u - 1

			} else {
				nSamples++
			}

			// reinitialise variables
			nMsgs = 0
			ticker.Reset(time.Duration(conf.Experiment.MonitorInterval) * time.Second)
		default:
		}
		// receive message
		message, err := receivers[currentReceiver].Receive(ctx, nil)
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				log.Print("sub stopped")
				return
			}
			log.Fatalf("receive message: %v", err)
		}

		// process message
		//log.Printf("received: %s", string(message.GetData()))
		//time.Sleep(ProcessingTime)

		// ack message
		if err := receivers[currentReceiver].AcceptMessage(ctx, message); err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return
			}
			log.Fatalf("accept message: %v", err)
		}
		nMsgs++
	}
}

func RunClosedLoop(conf config.Config) {

	// create context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// create receivers
	receivers, err := CreateReceivers(conf, ctx)
	if err != nil {
		log.Fatalf("Create receivers: %v", err)
		os.Exit(0)
	}

	// configure MICO
	m := mico.MICO{}
	err = m.Initialise(conf)
	if err != nil {
		fmt.Printf("Initialise mico: %v\n", err)
		os.Exit(1)
	}

	// initialisations
	ticker := time.NewTicker(time.Duration(conf.Experiment.MonitorInterval) * time.Second)
	nMsgs := 0
	nSamples := 0
	nSamplesPerLevel := 0
	contLevels := 0
	u := float64(conf.Experiment.InitialLinkedCredit)

	// loop
	fmt.Printf("************* Begin of Experiment [%v, %v, %v] ************* \n", conf.Experiment.ControllerType, conf.Experiment.ExecutionType, conf.RabbitMQ.QueueName)
	for {
		select {
		case <-ticker.C:
			// compute receiving rate
			r := float64(nMsgs) / float64(time.Duration(conf.Experiment.MonitorInterval))

			fmt.Printf("%v;%v;%v;%v\n", nSamples+1, int32(u), conf.Experiment.Setpoints[contLevels], int(r))

			// invoke controller
			u = m.Update(float64(conf.Experiment.Setpoints[contLevels]), r)
			conf.Experiment.CurrentReceiver = int(u) - 1

			// update & check samples
			nMsgs = 0
			nSamples++
			nSamplesPerLevel++

			if conf.Experiment.ExecutionType == shared.MultiClosedLoop {
				// check end of level
				if nSamplesPerLevel >= conf.Experiment.SamplesPerLevel {
					nSamplesPerLevel = 0
					contLevels++
				}

				// check end of experiment
				if contLevels >= len(conf.Experiment.Setpoints) {
					fmt.Printf("************* End of Experiment [%v, %v, %v] ************* \n", conf.Experiment.ControllerType, conf.Experiment.ExecutionType, conf.RabbitMQ.QueueName)
					os.Exit(1)
				}
			}
			if conf.Experiment.ExecutionType == shared.ClosedLoop {
				contLevels = contLevels
				// check end of experiments
				if nSamples >= conf.Experiment.SampleSize {
					fmt.Printf("************* End of Experiment [%v, %v, %v] ************* \n", conf.Experiment.ControllerType, conf.Experiment.ExecutionType, conf.RabbitMQ.QueueName)
					os.Exit(0)
				}
			}

			// reinitialise variables
			ticker.Reset(time.Duration(conf.Experiment.MonitorInterval) * time.Second)
		default:
		}

		// receive message
		message, err := receivers[conf.Experiment.CurrentReceiver].Receive(ctx, nil)
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				log.Print("sub stopped")
				return
			}
			log.Fatalf("receive message: %v", err)
		}

		// process message
		//log.Printf("received: %s", string(message.GetData()))
		//time.Sleep(ProcessingTime)

		// ack message
		if err := receivers[conf.Experiment.CurrentReceiver].AcceptMessage(ctx, message); err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return
			}
			log.Fatalf("accept message: %v", err)
		}
		nMsgs++
	}
}

func EnsureQueue(
	ctx context.Context,
	managementURL string,
	username string,
	password string,
	vhost string,
	queue string,
) error {
	if managementURL == "" {
		return errors.New("management URL must not be empty")
	}
	if vhost == "" {
		return errors.New("virtual host must not be empty")
	}
	if queue == "" {
		return errors.New("queue name must not be empty")
	}

	payload, err := json.Marshal(struct {
		AutoDelete bool           `json:"auto_delete"`
		Durable    bool           `json:"durable"`
		Arguments  map[string]any `json:"arguments"`
	}{
		AutoDelete: false,
		Durable:    true,
		Arguments:  map[string]any{"x-queue-ctype": "classic"},
	})
	if err != nil {
		return fmt.Errorf("encode queue declaration: %w", err)
	}

	endpoint := strings.TrimRight(managementURL, "/") +
		"/api/queues/" + url.PathEscape(vhost) + "/" + url.PathEscape(queue)
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		endpoint,
		bytes.NewReader(payload),
	)
	if err != nil {
		return fmt.Errorf("create queue declaration request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(username, password)

	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("execute queue declaration request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusOK &&
		response.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	body, readErr := io.ReadAll(io.LimitReader(response.Body, 8<<10))
	if readErr != nil {
		return fmt.Errorf(
			"queue declaration returned %s; response body could not be read: %w",
			response.Status,
			readErr,
		)
	}
	return fmt.Errorf(
		"queue declaration returned %s: %s",
		response.Status,
		strings.TrimSpace(string(body)),
	)
}

func CreateReceivers(conf config.Config, ctx context.Context) ([]amqp.Receiver, error) {
	// set of receivers
	receivers := []amqp.Receiver{}

	// create queue
	queueAddress := "/queues/" + url.PathEscape(conf.RabbitMQ.QueueName)
	err := EnsureQueue(
		ctx,
		"http://"+conf.RabbitMQ.Host+":"+strconv.Itoa(conf.RabbitMQ.PortManagement), //"http://localhost:15672",
		conf.RabbitMQ.User, //"guest",
		conf.RabbitMQ.Pwd,  //"guest",
		"/",
		conf.RabbitMQ.QueueName,
	)
	if err != nil {
		return nil, fmt.Errorf("declare queue: %v", err)
	}

	// create connection
	connection, err := amqp.Dial(
		ctx,
		"amqp:"+conf.RabbitMQ.Host+":"+strconv.Itoa(conf.RabbitMQ.Port), //amqp:localhost:5672",
		&amqp.ConnOptions{
			SASLType: amqp.SASLTypePlain(conf.RabbitMQ.User, conf.RabbitMQ.Pwd),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("connect to RabbitMQ: %v", err)
	}
	//defer connection.Close()

	// create pool of receivers
	for i := 0; i < conf.Experiment.MaxLinkCredit; i++ {
		// create session
		session, err := connection.NewSession(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("create session: %v", err)
		}
		//defer session.Close(context.Background())

		temp, err := session.NewReceiver(
			ctx,
			queueAddress,
			&amqp.ReceiverOptions{
				Credit: int32(i + 1),
			},
		)
		if err != nil {
			return nil, fmt.Errorf("create receiver: %v", err)
		}
		receivers = append(receivers, *temp)
		//defer receivers[i].Close(context.Background())
	}
	return receivers, err
}
