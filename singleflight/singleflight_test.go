package singleflight

import (
	"log"
	"sync"
	"testing"
	"time"
)

var wg sync.WaitGroup

func TestSF(t *testing.T) {
	group := Group{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go group.Do("test", func() (interface{}, error) {
			defer wg.Add(-100) //once
			log.Println("sleep.....")
			time.Sleep(time.Second * 1)
			log.Println("awake.....")
			return nil, nil
		})

	}

	wg.Wait()

}
