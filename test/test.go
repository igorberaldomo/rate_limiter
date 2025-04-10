package main

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)


func main() {
	bulkRequest := 15

	test("", bulkRequest)
	time.Sleep(2 * time.Second)
	test("token", bulkRequest)

}

func test(AuthToken string, bulkRequest int) {
	
	if AuthToken == "" {
		c := 0
		url, err := url.Parse("http://localhost:8080")
		if err != nil {
			panic(err)
		}
		name := "ip"
		ch := make(chan int, bulkRequest)
		header := http.Header{}
		header.Set("AuthToken", name)
		var wg sync.WaitGroup
		wg.Add(bulkRequest)
		start := time.Now()

		for range bulkRequest {
			go func() {

				req, err := http.NewRequest("GET", url.String(), nil)
				if err != nil {
					panic(err)
				}
				req.Header = header
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()
				ch <- resp.StatusCode
				c++
				wg.Done()
				if <- ch == http.StatusOK {
					fmt.Printf("request nº %d with ID alocado com sucesso\n", c)
				}
				if <- ch == http.StatusTooManyRequests {
					fmt.Printf("request nº %d with ID blockeado\n", c)
				} 
			}()
		}
		wg.Wait()
		totalTime := time.Since(start)
		fmt.Printf("Total time: %s\n", totalTime)


	}
	if AuthToken != "" {
		c := 0
		url, err := url.Parse("http://localhost:8080/" + AuthToken)
		if err != nil {
			panic(err)
		}
		ch := make(chan int, bulkRequest)
		header := http.Header{}
		header.Set("AuthToken", AuthToken)
		var wg sync.WaitGroup
		wg.Add(bulkRequest)
		start := time.Now()

		for range bulkRequest {
			go func() {
				req, err := http.NewRequest("GET", url.String(), nil)
				if err != nil {
					panic(err)
				}
				req.Header = header
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()
				ch <- resp.StatusCode
				c++

				wg.Done()
				if <- ch == http.StatusOK {
					fmt.Printf("request nº %d with token alocado com sucesso\n", c)
				}
				if <- ch == http.StatusTooManyRequests {
					fmt.Printf("request nº %d with token blockeado\n", c)
				} 
			}()
		}
		wg.Wait()
		totalTime := time.Since(start)
	
		fmt.Printf("Time: %s\n", totalTime.String())
		fmt.Print("\n")

	}

}
