package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	CheckUrl         string
	CheckTimeout     int64
	CheckInterval    int64
	RetryCount       int64
	StopCmd          string
	StartCmd         string
	WaitAfterStop    int64
	WaitAfterRestart int64
	ErrorCounter     int64
)

func init() {
	cpus := runtime.NumCPU()
	runtime.GOMAXPROCS(cpus)

	var err error
	log.SetFlags(log.Lshortfile)

	ErrorCounter = 0

	CheckUrl = os.Getenv("CHECK_URL")
	if len(CheckUrl) == 0 {
		log.Fatal(`Please set environment variable "CHECK_URL"`)
	}
	_, err = url.Parse(CheckUrl)
	if err != nil {
		log.Fatal(`Please set environment variable "CHECK_URL" as URL`)
	}
	log.Println(fmt.Sprintf(`CHECK_URL=[%s]`, CheckUrl))

	timeout := os.Getenv("CHECK_TIMEOUT")
	if len(timeout) == 0 {
		log.Fatal(`Please set environment variable "CHECK_TIMEOUT"`)
	}
	CheckTimeout, err = strconv.ParseInt(timeout, 10, 64)
	if err != nil {
		log.Fatal(`Please set environment variable "CHECK_TIMEOUT" as NUMERIC`)
	}
	log.Println(fmt.Sprintf(`CHECK_TIMEOUT=[%d]`, CheckTimeout))

	interval := os.Getenv("CHECK_INTERVAL")
	if len(interval) == 0 {
		log.Fatal(`Please set environment variable "CHECK_INTERVAL"`)
	}
	CheckInterval, err = strconv.ParseInt(interval, 10, 64)
	if err != nil {
		log.Fatal(`Please set environment variable "CHECK_INTERVAL" as NUMERIC`)
	}
	log.Println(fmt.Sprintf(`CHECK_INTERVAL=[%d]`, CheckInterval))

	count := os.Getenv("RETRY_COUNT")
	if len(count) == 0 {
		log.Fatal(`Please set environment variable "RETRY_COUNT"`)
	}
	RetryCount, err = strconv.ParseInt(count, 10, 64)
	if err != nil {
		log.Fatal(`Please set environment variable "RETRY_COUNT" as NUMERIC`)
	}
	log.Println(fmt.Sprintf(`RETRY_COUNT=[%d]`, RetryCount))

	StopCmd = os.Getenv("STOP_COMMAND")
	if len(StopCmd) == 0 {
		log.Fatal(`Please set environment variable "STOP_COMMAND"`)
	}
	log.Println(fmt.Sprintf(`STOP_COMMAND=[%s]`, StopCmd))

	StartCmd = os.Getenv("START_COMMAND")
	if len(StartCmd) == 0 {
		log.Fatal(`Please set environment variable "START_COMMAND"`)
	}
	log.Println(fmt.Sprintf(`START_COMMAND=[%s]`, StartCmd))

	wait := os.Getenv("WAIT_AFTER_STOP")
	if len(wait) == 0 {
		log.Fatal(`Please set environment variable "WAIT_AFTER_STOP"`)
	}
	WaitAfterStop, err = strconv.ParseInt(wait, 10, 64)
	if err != nil {
		log.Fatal(`Please set environment variable "WAIT_AFTER_STOP" as NUMERIC`)
	}
	log.Println(fmt.Sprintf(`WAIT_AFTER_STOP=[%d]`, WaitAfterStop))

	wait = os.Getenv("WAIT_AFTER_RESTART")
	if len(wait) == 0 {
		log.Fatal(`Please set environment variable "WAIT_AFTER_RESTART"`)
	}
	WaitAfterRestart, err = strconv.ParseInt(wait, 10, 64)
	if err != nil {
		log.Fatal(`Please set environment variable "WAIT_AFTER_RESTART" as NUMERIC`)
	}
	log.Println(fmt.Sprintf(`WAIT_AFTER_RESTART=[%d]`, WaitAfterRestart))

	http.DefaultClient.Timeout = time.Second * time.Duration(CheckTimeout)
}

func check() {
	c := Check{
		client: http.DefaultClient,
	}

	req, err := http.NewRequest(http.MethodGet, CheckUrl, nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	defer func() {
		if ErrorCounter < RetryCount {
			return
		}

		err := c.restart()
		if err != nil {
			log.Fatal(err)
		}
		ErrorCounter = 0
		time.Sleep(time.Second * time.Duration(WaitAfterRestart))
	}()

	res, err := c.request(context.Background(), req)
	if err != nil {
		log.Println(err)
		ErrorCounter++
		return
	}

	if res.StatusCode >= 400 {
		log.Println(fmt.Sprintf("bad response status code %d", res.StatusCode))
		ErrorCounter++
		return
	}
	log.Println(fmt.Sprintf("%s %d", CheckUrl, res.StatusCode))
	ErrorCounter = 0
}

type Check struct {
	client *http.Client
}

func (c *Check) request(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)

	resCh := make(chan *http.Response)
	errCh := make(chan error)

	go func() {
		res, err := c.client.Do(req)
		if err != nil {
			errCh <- err
			return
		}

		resCh <- res
	}()

	select {
	case res := <-resCh:
		return res, nil

	case err := <-errCh:
		return nil, err

	case <-ctx.Done():
		return nil, errors.New("HTTP request cancelled")
	}
}

func (c *Check) restart() error {
	var cmd []string
	var err error

	log.Println(`restart...`)
	cmd = strings.Fields(StopCmd)
	stop := exec.Command(cmd[0], cmd[1:]...)
	log.Println(fmt.Sprintf(`exec stop command. [%#v]`, cmd))

	err = stop.Run()
	if err != nil {
		return err
	}

	time.Sleep(time.Second * time.Duration(WaitAfterStop))

	cmd = strings.Fields(StartCmd)
	start := exec.Command(cmd[0], cmd[1:]...)
	log.Println(fmt.Sprintf(`exec start command. [%#v]`, cmd))

	err = start.Run()
	if err != nil {
		return err
	}
	log.Println(`done...`)
	return nil
}

func main() {
	t := time.NewTicker(time.Second * time.Duration(CheckInterval))
	for {
		select {
		case <-t.C:
			check()
		}
	}
}
