package main

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)


func main() {
	bulkRequest := 10

	test("", bulkRequest)
	test("token", bulkRequest)

}

func test(AuthToken string, bulkRequest int) {
	if AuthToken == "" {
		url, err := url.Parse("http://localhost:8080")
		if err != nil {
			panic(err)
		}
		AuthToken = "ip"
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
				wg.Done()
				fmt.Print("test with IP\n")
				fmt.Printf("User: %s\n", AuthToken)
				fmt.Printf("Status: %d\n", <-ch)
				fmt.Print("-----------------\n")
			}()
		}
		wg.Wait()
		totalTime := time.Since(start)
		fmt.Printf("Total time: %s\n", totalTime)


	}
	if AuthToken != "" {
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
				wg.Done()
				fmt.Print("test with Auth\n")
				fmt.Printf("User: %s\n", AuthToken)
				fmt.Printf("Status: %d\n", <-ch)
				fmt.Print("-----------------\n")
		
			}()
		}
		wg.Wait()
		totalTime := time.Since(start)
	
		fmt.Printf("Time: %s\n", totalTime.String())
		fmt.Print("\n")

	}

}
