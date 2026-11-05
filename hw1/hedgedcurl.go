package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Result struct {
	resp *http.Response
	err  error
}

func makeRequest(ctx context.Context, client *http.Client, url string, results chan<- Result) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		results <- Result{err: err}
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		results <- Result{err: err}
		return
	}
	results <- Result{resp: resp}
}


func main() {

	var help bool

	flag.Bool(&help,"help", false, "Show help")
	flag.Bool(&help,"h", false, "Show help")
	
	var t int
	flag.IntVar(&t, "t", 15, "Timer")
	flag.IntVar(&t, "timeout", 15, "Timer")

	flag.Parse()

	if help {
		fmt.Println("Использование: hedgedcurl [-t,--timeout seconds] URL...")
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Error: provide at least one URL")
		os.Exit(1)
	}

	if *t <= 0 {
		fmt.Println("Error: timeout must be positive")
		os.Exit(1)
	}

	client := &http.Client{}
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(*t)*time.Second,
	)
	defer cancel()

	results := make(chan Result, len(args))
	for _, url := range args {
		go makeRequest(ctx, client, url, results)
	}

	gotResponse := false
	var lastErr error

	for i := 0; i < len(args); i++ {
		res := <-results
		if res.err != nil{
			lastErr = res.err
			continue
		}

		gotResponse = true

		fmt.Println(res.resp.Status)
		for name, values := range res.resp.Header {
			for _, val := range values {
				fmt.Printf("%s: %s\n", name, val)
			}
		}
		fmt.Println()

		_, err := io.Copy(os.Stdout, res.resp.Body)

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading body:", err)

			if errors.Is(err, context.DeadlineExceeded) {
				res.resp.Body.Close()
				cancel()
				os.Exit(228)
			}

			res.resp.Body.Close()
			cancel()
			os.Exit(1)
		}

		res.resp.Body.Close()
		cancel()
		break

	}

	if !gotResponse {
		fmt.Fprintln(os.Stderr, "Error: all requests failed:", lastErr)
		if ctx.Err() == context.DeadlineExceeded {
			os.Exit(228)
		}
		os.Exit(1)
	}
}
