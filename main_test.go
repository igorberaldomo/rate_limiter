package main

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

func init(){
	bulkRequest := 1000

	test("", bulkRequest)
	test("token", bulkRequest)
	
}

func test(AuthToken string, bulkRequest int) {
	if AuthToken == "" {
		url, err := url.Parse("http://localhost:8080")
		if err != nil {
			panic(err)
		}
		ch := make(chan int,bulkRequest)
		header := http.Header{}
		header.Set("AuthToken", AuthToken)
		var wg sync.WaitGroup
		wg.Add(bulkRequest)
		start := time.Now()
	
		for range bulkRequest{
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
			}()
		}
		wg.Wait()
		totalTime := time.Since(start)
		fmt.Printf("User:" + AuthToken + "\n")  
		fmt.Printf("Time:" + totalTime.String()+ "\n")
		fmt.Printf("Status:" + fmt.Sprint(ch) + "\n")
		
	}
	if AuthToken != "" {
		url, err := url.Parse("http://localhost:8080/?AuthToken=" + AuthToken)
		if err != nil {
			panic(err)
		}
		ch := make(chan int,bulkRequest)
		header := http.Header{}
		header.Set("AuthToken", AuthToken)
		var wg sync.WaitGroup
		wg.Add(bulkRequest)
		start := time.Now()
	
		for range bulkRequest{
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
			}()
		}
		wg.Wait()
		totalTime := time.Since(start)
		fmt.Printf("User:" + AuthToken + "\n")  
		fmt.Printf("Time:" + totalTime.String()+ "\n")
		fmt.Printf("Status:" + fmt.Sprint(ch) + "\n")

	}



}