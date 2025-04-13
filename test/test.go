package main

import (
	"fmt"
	"net/http"
	"net/url"
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
		url, err := url.Parse("http://localhost:8080")
		if err != nil {
			panic(err)
		}
		header := http.Header{}
		header.Set("AuthToken", "ip")
		start := time.Now()
		for i := 0; i < bulkRequest; i++ {
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

			result := resp.StatusCode
			if result == http.StatusOK {
				fmt.Printf("request nº %d with ID alocado com sucesso\n", i+1)
			}
			if result == http.StatusTooManyRequests {
				fmt.Printf("request nº %d with ID blockeado\n", i+1)
			} 
		}
		totalTime := time.Since(start)
		fmt.Printf("Total time: %s\n", totalTime)
	}
	if AuthToken != "" {

		url, err := url.Parse("http://localhost:8080/" + AuthToken)
		if err != nil {
			panic(err)
		}
		header := http.Header{}
		header.Set("AuthToken", AuthToken)
		start := time.Now()
		for i := 0; i < bulkRequest; i++ {
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

				result := resp.StatusCode
				
				if result == http.StatusOK {
					fmt.Printf("request nº %d with token alocado com sucesso\n", i+1)
				}
				if result == http.StatusTooManyRequests {
					fmt.Printf("request nº %d with token blockeado\n", i+1)
				} 
		}
		totalTime := time.Since(start)
	
		fmt.Printf("Time: %s\n", totalTime.String())
		fmt.Print("\n")

	}

}
