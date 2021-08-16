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

// producer
func producer(channel chan bool) {

	resources := createRandom(1, 3)

	for i := 0; i < resources; i++ {
		// as a dummy resource, boolean value true is used
		channel <- true
	}

	fmt.Println("produced ", resources)

}

// consumer
func consumer(producerChannels []chan bool, wg *sync.WaitGroup) {

	// waitGroup for consumer to finish
	defer wg.Done()

	// count number of resources consumed
	storage := 0

	// run for 10 batch
	for batch := 0; batch < 10; batch++ {

		producerIndex := 0

		for {

			if storage == 10 {

				sleepTime := createRandom(1, 5)
				fmt.Println("discarded")
				time.Sleep(time.Duration(sleepTime) * time.Second)
				storage = 0
				break
			}

			getNum := createRandom(1, 3)
			// rotate through the four producers

			consumed := 0
			for consumed < getNum && storage < 10 {
				producerIndex %= 3

				select {

				case <-producerChannels[producerIndex]:
					// values are available in the producer

					consumed++
					storage++

					var channelCleared bool
					for consumed < getNum && storage < 10 && !channelCleared {

						select {

						case <-producerChannels[producerIndex]:
							// read one value
							consumed++
							storage++

						default:
							// finished reading from producer, move to next producer
							producerIndex++
							channelCleared = true
						}
					}

				default:
					// producer is empty, awake producer
					go producer(producerChannels[producerIndex])
				}

			}
			fmt.Println("consumed ", consumed)

		}
	}
}

func main() {

	fmt.Println("started")

	producerChannels := []chan bool{
		make(chan bool, 10),
		make(chan bool, 10),
		make(chan bool, 10),
		make(chan bool, 10),
	}

	go producer(producerChannels[0])
	go producer(producerChannels[1])
	go producer(producerChannels[2])
	go producer(producerChannels[3])

	var wg sync.WaitGroup

	wg.Add(2)

	go consumer(producerChannels, &wg)
	go consumer(producerChannels, &wg)

	wg.Wait()

}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
