package main

import (
	"adaptivesystemscourse/internal/config"
	"adaptivesystemscourse/internal/control/mico"
	"adaptivesystemscourse/internal/shared"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"syscall"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

func main() {
	// load config
	conf := shared.LoadConfig("subscriber")

	// configure some variables
	conf.Experiment.CurrentReceiver = conf.Experiment.InitialLinkedCredit
	conf.Experiment.HysteresisBand = conf.Experiment.Setpoint * 0.05

	switch conf.Experiment.ExecutionType {
	case shared.ClosedLoop:
		RunClosedLoop(*conf)
	case shared.MultiClosedLoop:
		//RunMultiClosedLoop(*conf)
		RunClosedLoop(*conf)
	case shared.OpenLoop:
		RunOpenLoop(*conf)
	default:
		fmt.Println("execution type unknown")
		os.Exit(1)
	}
	fmt.Scanln()
}

/*
func RunClosedLoop(conf config.Config) {

		// create sub & actuator
		actuator, consumer, err := CreateConsumer(conf)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		// configure initial rate goal
		dispatchRate := int64(conf.Experiment.InitialDispatchRate) // xx msg/s
		if err := actuator.SetDispatchRate(ctx, int64(conf.Experiment.InitialDispatchRate)); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		// configure MICO
		m := mico.MICO{}
		err = m.Initialise(conf)
		if err != nil {
			log.Fatalf("Error in initialising MICO: %v", err)
			os.Exit(1)
		}

		// initialise counters
		nMsgs := 0
		nSamples := 0
		ticker := time.NewTicker(time.Duration(conf.Experiment.MonitorInterval) * time.Second)

		fmt.Printf("****** Begin of Experiment [%v, %v, %v]******\n", conf.Experiment.ExecutionType, conf.Experiment.ControllerType, conf.Pulsar.TopicName)
		for {
			select {
			case <-ticker.C:

				// compute actual rate
				actualRate := float64(nMsgs) / float64(conf.Experiment.MonitorInterval) // seconds
				fmt.Printf("%v;%v;%v;%v\n", nSamples+1, dispatchRate, conf.Experiment.Setpoint, int(actualRate))

				// invoke controller
				dispatchRate = int64(m.Update(conf.Experiment.Setpoint, actualRate))

				// adjust the dispatch rate
				if err := actuator.SetDispatchRate(ctx, dispatchRate); err != nil {
					log.Fatalf("change dispatch rate: %v", err)
				}

				// update counters
				nSamples++

				// check end of experiments
				if nSamples >= conf.Experiment.SampleSize {
					fmt.Printf("************* End of Experiment [%v, %v, %v] ************* \n", conf.Experiment.ControllerType, conf.Experiment.ExecutionType, conf.RabbitMQ.QueueName)
					os.Exit(1)
				}

				// reset counters
				nMsgs = 0
				ticker.Reset(time.Duration(conf.Experiment.MonitorInterval) * time.Second)
			default:
			}
			// receive message
			message, err := consumer.Receive(ctx)
			if err != nil {
				log.Fatalf("receive message: %v", err)
			}

			// send ack back to the broker
			consumer.Ack(message)

			// increment message counter
			nMsgs++
		}
	}
*/
func RunClosedLoop(conf config.Config) {

	// create sub & actuator
	actuator, consumer, err := CreateConsumer(conf)
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// configure initial rate goal
	dispatchRate := int64(conf.Experiment.InitialDispatchRate) // xx msg/s
	if err := actuator.SetDispatchRate(ctx, int64(conf.Experiment.InitialDispatchRate)); err != nil {
		log.Fatalf("set dispatch rate: %v", err)
	}

	// configure MICO
	m := mico.MICO{}
	err = m.Initialise(conf)
	if err != nil {
		log.Fatalf("Error in initialising MICO: %v", err)
		os.Exit(0)
	}

	// initialise counters
	nMsgs := 0
	nSamples := 0
	nSamplesPerLevel := 0
	contLevels := 0
	ticker := time.NewTicker(time.Duration(conf.Experiment.MonitorInterval) * time.Second)

	fmt.Printf("****** Begin of Experiment [%v, %v, %v]******\n", conf.Experiment.ExecutionType, conf.Experiment.ControllerType, conf.Pulsar.TopicName)
	for {
		select {
		case <-ticker.C:

			// compute actual rate
			actualRate := float64(nMsgs) / float64(time.Duration(conf.Experiment.MonitorInterval))
			fmt.Printf("%v;%v;%v;%v\n", nSamples+1, dispatchRate, conf.Experiment.Setpoints[contLevels], int(actualRate))

			// invoke controller
			dispatchRate = int64(m.Update(conf.Experiment.Setpoints[contLevels], actualRate))

			// adjust the dispatch rate
			if err := actuator.SetDispatchRate(ctx, dispatchRate); err != nil {
				log.Fatalf("change dispatch rate: %v", err)
			}

			// update counters
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
					fmt.Printf("****** End of Experiment [%v, %v, %v]******\n", conf.Experiment.ExecutionType, conf.Experiment.ControllerType, conf.Pulsar.TopicName)
					os.Exit(1)
				}
			}
			if conf.Experiment.ExecutionType == shared.ClosedLoop {
				contLevels = contLevels
				// check end of experiments
				if nSamples >= conf.Experiment.SampleSize {
					fmt.Println("****** End of Experiment ******")
					os.Exit(1)
				}
			}

			// reset counters
			nMsgs = 0
			ticker.Reset(time.Duration(conf.Experiment.MonitorInterval) * time.Second)
		default:
		}
		// receive message
		message, err := consumer.Receive(ctx)
		if err != nil {
			log.Fatalf("receive message: %v", err)
		}

		// send ack back to the broker
		consumer.Ack(message)

		// increment message counter
		nMsgs++
	}
}

func RunOpenLoop(conf config.Config) {

	// create sub & actuator
	actuator, consumer, err := CreateConsumer(conf)
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// configure initial distpatch rate
	dispatchRate := int64(conf.Experiment.InitialDispatchRate) // xx msg/s
	if err := actuator.SetDispatchRate(ctx, int64(conf.Experiment.InitialDispatchRate)); err != nil {
		log.Fatalf("set dispatch rate: %v", err)
	}

	// configure MICO
	m := mico.MICO{}
	err = m.Initialise(conf)
	if err != nil {
		log.Fatalf("Error in initialising MICO: %v", err)
		os.Exit(0)
	}

	// initialise counters
	nMsgs := 0
	nSamples := 0
	ticker := time.NewTicker(time.Duration(conf.Experiment.MonitorInterval))

	fmt.Println("****** Begin of Experiment ******")

	for {
		select {
		case <-ticker.C:

			// compute actual rate
			r := float64(nMsgs) / float64(time.Duration(conf.Experiment.MonitorInterval).Seconds())
			fmt.Printf("%v;%v;%v\n", time.Now(), dispatchRate, r)

			// update dispatch rate
			if nSamples >= conf.Experiment.SampleSize-1 {
				nSamples = 0
				dispatchRate += 10

				// adjust dispatch rate
				if err := actuator.SetDispatchRate(ctx, dispatchRate); err != nil {
					log.Fatalf("change dispatch rate: %v", err)
				}
			} else {
				nSamples++
			}

			// reset counters
			nMsgs = 0
			ticker.Reset(time.Duration(conf.Experiment.MonitorInterval))
		default:
		}

		// receive message
		message, err := consumer.Receive(ctx)
		if err != nil {
			log.Fatalf("receive message: %v", err)
		}

		// send ack back to broker
		consumer.Ack(message)

		// increment message counter
		nMsgs++
	}
}

func percentileDuration(
	sample []time.Duration,
	percentile float64,
) (time.Duration, error) {
	if len(sample) == 0 {
		return 0, errors.New("sample cannot be empty")
	}

	if percentile <= 0 || percentile > 100 {
		return 0, errors.New(
			"percentile must be greater than 0 and at most 100",
		)
	}

	// Copy the sample so the original slice is not modified.
	sorted := append([]time.Duration(nil), sample...)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	// Nearest-rank method.
	rank := int(math.Ceil(
		percentile / 100.0 * float64(len(sorted)),
	))

	for i := range sorted {
		fmt.Println(sorted[i].Milliseconds(), sorted[rank].Milliseconds())
	}
	return sorted[rank-1], nil
}

type TopicStats struct {
	Subscriptions map[string]SubscriptionStats `json:"subscriptions"`
}

type SubscriptionStats struct {
	MsgBacklog                      int64   `json:"msgBacklog"`
	BacklogSize                     int64   `json:"backlogSize"`
	MsgBacklogNoDelayed             int64   `json:"msgBacklogNoDelayed"`
	UnackedMessages                 int64   `json:"unackedMessages"`
	MsgDelayed                      int64   `json:"msgDelayed"`
	MsgRateOut                      float64 `json:"msgRateOut"`
	MsgThroughputOut                float64 `json:"msgThroughputOut"`
	MessageAckRate                  float64 `json:"messageAckRate"`
	EarliestMsgPublishTimeInBacklog int64   `json:"earliestMsgPublishTimeInBacklog"`
}

type PulsarAdminClient struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func (p *PulsarAdminClient) GetSubscriptionStats(
	ctx context.Context,
	tenant string,
	namespace string,
	topic string,
	subscription string,
	precise bool,
) (SubscriptionStats, error) {
	endpoint := fmt.Sprintf(
		"%s/admin/v2/persistent/%s/%s/%s/stats",
		p.BaseURL,
		url.PathEscape(tenant),
		url.PathEscape(namespace),
		url.PathEscape(topic),
	)

	// Supported by current Pulsar Admin REST implementations.
	if precise {
		endpoint += "?get_precise_backlog=true"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return SubscriptionStats{}, err
	}

	if p.Token != "" {
		req.Header.Set("Authorization", "Bearer "+p.Token)
	}

	client := p.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return SubscriptionStats{}, fmt.Errorf("request topic stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return SubscriptionStats{}, fmt.Errorf(
			"pulsar admin returned %s: %s",
			resp.Status,
			string(body),
		)
	}

	var stats TopicStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return SubscriptionStats{}, fmt.Errorf("decode topic stats: %w", err)
	}

	subStats, found := stats.Subscriptions[subscription]
	if !found {
		return SubscriptionStats{}, fmt.Errorf(
			"subscription %q not found",
			subscription,
		)
	}

	return subStats, nil
}

// DurationPercentile calculates the percentile of a time.Duration slice using linear interpolation.
func durationPercentile(input []time.Duration, percent float64) (time.Duration, error) {
	if len(input) == 0 {
		return 0, fmt.Errorf("input slice cannot be empty")
	}
	if percent < 0 || percent > 100 {
		return 0, fmt.Errorf("percent must be between 0 and 100")
	}

	// 1. Duplicate the dataset so the original slice isn't mutated
	data := make([]time.Duration, len(input))
	copy(data, input)

	// 2. Sort the durations (time.Duration natively supports sorting using Slice or SliceStable)
	sort.Slice(data, func(i, j int) bool {
		return data[i] < data[j]
	})

	// Edge cases for exact boundaries
	if percent == 0 {
		return data[0], nil
	}
	if percent == 100 {
		return data[len(data)-1], nil
	}

	// 3. Compute the rank/index
	rank := (percent / 100) * float64(len(data)-1)

	// 4. Split into integer and fractional components
	k := int(math.Floor(rank))
	f := rank - float64(k)

	// 5. Interpolate using nanosecond values, then convert back to time.Duration
	d1 := float64(data[k].Nanoseconds())
	d2 := float64(data[k+1].Nanoseconds())
	interpolatedNs := d1 + f*(d2-d1)

	// Round to nearest integer nanosecond before casting back to time.Duration
	return time.Duration(math.Round(interpolatedNs)), nil
}

func CreateConsumer(conf config.Config) (shared.DispatchRateActuator, pulsar.Consumer, error) {
	//serviceURL := flag.String("url", "pulsar://localhost:6650", "Pulsar service URL")
	//flag.Parse()

	// configuration
	adminURL := flag.String(
		"admin-url",
		"http://"+conf.Pulsar.Host+":"+strconv.Itoa(conf.Pulsar.PortManagement),
		"Pulsar Admin API URL",
	)
	tenant := flag.String(
		"tenant",
		"public",
		"Pulsar tenant",
	)
	namespace := flag.String(
		"namespace",
		"default",
		"Pulsar namespace",
	)
	token := flag.String(
		"token",
		"",
		"optional authentication token",
	)
	flag.Parse()

	actuator := shared.DispatchRateActuator{
		AdminURL:  *adminURL,
		Tenant:    *tenant,
		Namespace: *namespace,
		Token:     *token,
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	// create client
	client, err := pulsar.NewClient(pulsar.ClientOptions{URL: "pulsar://" + conf.Pulsar.Host + ":" + strconv.Itoa(conf.Pulsar.Port)})
	if err != nil {
		return actuator, nil, fmt.Errorf("create client: %v", err)
	}
	//defer client.Close()

	// create subscriber
	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:             conf.Pulsar.TopicName,
		SubscriptionName:  conf.Pulsar.SubscriptionName,
		Type:              pulsar.Shared,
		ReceiverQueueSize: conf.Pulsar.ReceiverQueueSize,
	})
	if err != nil {
		return actuator, consumer, fmt.Errorf("subscribe: %v", err)
	}

	return actuator, consumer, err
}
