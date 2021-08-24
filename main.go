package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// rand.Intn only accepts single value n and returns value between [0, n). following line converts it into [min, max] range
// source : https://stackoverflow.com/a/23577092
func createRandom(min, max int) int {

	return rand.Intn(max-min+1) + 1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Function to consume values from a producer is resources are available
func consumeFromProducer(producerData *chan bool, consumeLimit int) (int, bool) {

	var consumed int

	// read consumeLimit number of resources from the producer producerData
	for consumed <= consumeLimit {

		select {

		case <-*producerData:
			// values are available in the producer, then read the values

			consumed++

			for consumed < consumeLimit {

				select {

				case <-*producerData:
					// read one value
					consumed++

				default:

					// finished reading all values from producer, return number of resorces read and true flag for successfull read
					return consumed, true
				}
			}

			// succesfull consumed required num of resources, return number of resorces read and true flag for successfull read
			return consumed, true

		default:
			// if values are not available in the producer return false value

			return 0, false
			// producer is empty, awake producer
			//awakeSignals[producerIndex] <- true
		}
	}

	// after succesfull consumption, return the number of resorces consumed and true flag for successfull read
	return consumed, true
}

// producer
// in this design, once a producer is launched, it will contiue to produce resources when even it gets an awake signal
// in case of no awake signal, producer sleeps waiting for awake signal
func producer(dataChannel chan bool, awakeSignal chan bool) {

	// keep on producing resources in batches when awakeSignal is received
	for {

		select {

		// producer waits for awake signal to start producing
		case <-awakeSignal:
			// if awake signal is passed to producer

			resources := createRandom(1, 3)

			for i := 0; i < resources; i++ {
				// as a dummy resource, boolean value true is used
				// since the channel is buffered, producer will be blocked here when channel is full
				dataChannel <- true
			}

			fmt.Println("produced ", resources)

		}

	}
}

// consumer
func consumer(producerDataChannels []chan bool, awakeSignals []chan bool, wg *sync.WaitGroup) {

	// waitGroup for consumer to finish
	defer wg.Done()

	// count number of resources consumed
	storage := 0

	// run for 10 batch

	producerIndex := 0
	for batch := 0; batch < 10; batch++ {

		for {

			if storage == 10 {

				sleepTime := createRandom(1, 5)
				fmt.Println("discarded")
				time.Sleep(time.Duration(sleepTime) * time.Second)
				storage = 0
				break
			}

			// consume b/w 1 and 3 resources at a time
			getNum := createRandom(1, 3)
			consumed := 0

			// rotate through the four producers
			for storage < 10 && consumed < getNum {

				toConsume := min(getNum-consumed, 10-storage)

				if value, alive := consumeFromProducer(&producerDataChannels[producerIndex], toConsume); alive {
					// producer was alive and returned values
					storage += value
					consumed += value

				} else {
					// producer is empty, awake the producer and try again
					awakeSignals[producerIndex] <- true
					continue
				}

				// move to next producer
				producerIndex = (producerIndex + 1) % 3

			}

			fmt.Println("consumed ", consumed)

		}
	}
}

func main() {

	fmt.Println("started")

	producerDataChannels := []chan bool{
		make(chan bool, 10),
		make(chan bool, 10),
		make(chan bool, 10),
		make(chan bool, 10),
	}

	// awake signal to start producer
	awakeSignals := []chan bool{
		make(chan bool, 2),
		make(chan bool, 2),
		make(chan bool, 2),
		make(chan bool, 2),
	}

	// awake all producers by default once
	for _, channel := range awakeSignals {
		channel <- true
	}

	go producer(producerDataChannels[0], awakeSignals[0])
	go producer(producerDataChannels[1], awakeSignals[1])
	go producer(producerDataChannels[2], awakeSignals[2])
	go producer(producerDataChannels[3], awakeSignals[3])

	var wg sync.WaitGroup

	wg.Add(2)

	go consumer(producerDataChannels, awakeSignals, &wg)
	go consumer(producerDataChannels, awakeSignals, &wg)

	wg.Wait()

}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
