package main

import (
	"context"
	"flag"
	"fmt"
	"hash"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/G-V-G/2.l2/httptools"
	"github.com/G-V-G/2.l2/signal"
)

var (
	port         = flag.Int("port", 8090, "load balancer port")
	timeoutSec   = flag.Int("timeout-sec", 3, "request timeout time in seconds")
	https        = flag.Bool("https", false, "whether backends support HTTPs")
	traceEnabled = flag.Bool("trace", false, "whether to include tracing information into responses")
	timeout      = time.Duration(*timeoutSec) * time.Second
	serversPool  = []string{
		"server1:8080",
		"server2:8080",
		"server3:8080",
	}
)

func scheme() string {
	if *https {
		return "https"
	}
	return "http"
}

func health(dst string) bool {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	req, _ := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s://%s/health", scheme(), dst), nil)
	if resp, err := http.DefaultClient.Do(req); err != nil {
		log.Println(err.Error())
		return false
	} else {
		return resp.StatusCode == http.StatusOK
	}
}

func forward(dst string, rw http.ResponseWriter, r *http.Request) error {
	ctx, _ := context.WithTimeout(r.Context(), timeout)
	fwdRequest := r.Clone(ctx)
	fwdRequest.RequestURI = ""
	fwdRequest.URL.Host = dst
	fwdRequest.URL.Scheme = scheme()
	fwdRequest.Host = dst

	resp, err := http.DefaultClient.Do(fwdRequest)
	if err != nil {
		log.Printf("Failed to get response from %s: %s", dst, err)
		rw.WriteHeader(http.StatusServiceUnavailable)
		return err
	}

	for k, values := range resp.Header {
		for _, value := range values {
			rw.Header().Add(k, value)
		}
	}
	if *traceEnabled {
		rw.Header().Set("lb-from", dst)
	}
	log.Println("fwd", resp.StatusCode, resp.Request.URL)
	rw.WriteHeader(resp.StatusCode)
	defer resp.Body.Close()
	if _, err := io.Copy(rw, resp.Body); err != nil {
		log.Printf("Failed to write response: %s", err)
	}

	return nil
}

func hashPath(urlPath string) uint64 {
	var h64 hash.Hash64 = fnv.New64()
	h64.Write([]byte(urlPath))
	return h64.Sum64()
}

func balance(healthPool *HostsHealth, url string) (string, error) {
	healthyHosts := healthPool.GetHealthy()
	healthyAmount := len(healthyHosts)
	if healthyAmount == 0 {
		return "", fmt.Errorf("no servers available")
	}
	totalAmount := len(*healthPool)
	hashed := hashPath(url)
	idx := hashed % uint64(totalAmount)
	found := (*healthPool)[idx]
	if found.isHealthy {
		return found.addr, nil
	}
	idxHealthy := hashed % uint64(healthyAmount)
	return healthyHosts[idxHealthy], nil
}

func main() {
	flag.Parse()
	healthPool, _ := NewHealthChecker(&serversPool)

	// TODO: ???????????????????????????? ???????? ?????? ???????? ??????????????, ?????? ???????????????????????? ???????????? ?????? ????????????????, ???????? ?????????? ?????????????????????? ????????????.
	for i, _ := range serversPool {
		i := i
		go func() {
			for range time.Tick(10 * time.Second) {
				isHealthy := health(serversPool[i])
				healthPool.SetHealthState(i, isHealthy)
				log.Println(serversPool[i], isHealthy)
			}
		}()
	}

	frontend := httptools.CreateServer(*port, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// TODO: ?????????????????????? ???????? ???????????????? ??????????????????????????????.
		if server, err := balance(healthPool, r.URL.Path); err != nil {
			rw.WriteHeader(http.StatusServiceUnavailable)
			rw.Write([]byte(err.Error()))
		} else {
			forward(server, rw, r)
		}
	}))

	log.Println("Starting load balancer...")
	log.Printf("Tracing support enabled: %t", *traceEnabled)
	frontend.Start()
	signal.WaitForTerminationSignal()
}
