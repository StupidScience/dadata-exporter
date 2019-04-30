package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const (
	namespace = "dadata"
)

// structs for describe dadata api responses
type dadataBalance struct {
	Balance float64 `json:"balance"`
}

type dadataStats struct {
	Date     string `json:"date"`
	Services dadataServices
}

type dadataServices struct {
	Merging     int `json:"merging"`
	Suggestions int `json:"suggestions"`
	Clean       int `json:"clean"`
}

type dadataError struct {
	e          string
	statusCode int
}

func (d dadataError) Error() string {
	return fmt.Sprintf("%s, got status code %d", d.e, d.statusCode)
}

// Collector type for prometheus.Collector interface implementation
type Collector struct {
	CurrentBalance      prometheus.Gauge
	ServicesMerging     prometheus.Gauge
	ServicesSuggestions prometheus.Gauge
	ServicesClean       prometheus.Gauge

	totalScrapes         prometheus.Counter
	failedBalanceScrapes prometheus.Counter
	failedStatsScrapes   prometheus.Counter

	dadataAPIURL  string
	dadataToken   string
	dadataXSecret string

	sync.Mutex
}

// Describe for prometheus.Collector interface implementation
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

// Collect for prometheus.Collector interface implementation
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.Lock()
	defer c.Unlock()

	c.totalScrapes.Inc()
	err := c.getDadataBalance()
	if err != nil {
		c.failedBalanceScrapes.Inc()
	}
	err = c.getDadataStats()
	if err != nil {
		c.failedStatsScrapes.Inc()
	}

	ch <- c.totalScrapes
	ch <- c.failedBalanceScrapes
	ch <- c.failedStatsScrapes
	ch <- c.CurrentBalance
	ch <- c.ServicesClean
	ch <- c.ServicesMerging
	ch <- c.ServicesSuggestions
}

func (c *Collector) dadataRequest(m string) (*http.Response, error) {
	reqURL := fmt.Sprintf("%s/%s", c.dadataAPIURL, m)
	client := &http.Client{}
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		log.Errorf("Cannot create new request: %v", err)
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.dadataToken))
	req.Header.Add("X-Secret", c.dadataXSecret)

	r, err := client.Do(req)
	if err != nil {
		log.Errorf("Cannot make new request to %s: %v", reqURL, err)
		return nil, err
	}
	switch r.StatusCode {
	case http.StatusOK:
		return r, nil
	case http.StatusForbidden:
		return nil, fmt.Errorf("Can't access with provided Token/X-Secret pair")
	default:
		return nil, dadataError{
			e:          fmt.Sprintf("Error occurred on %s", reqURL),
			statusCode: r.StatusCode,
		}
	}

}

func (c *Collector) dadataCheck() error {
	_, err := c.dadataRequest("profile/balance")
	if err != nil {
		switch err.(type) {
		case dadataError:
			log.Errorf(err.Error())
			return err
		default:
			return err
		}

	}
	return nil
}

func (c *Collector) getDadataBalance() error {
	r, err := c.dadataRequest("profile/balance")
	if err != nil {
		log.Errorf("Cannot make request for balance: %v", err)
		return err
	}
	defer r.Body.Close()
	var b dadataBalance

	err = json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		log.Errorf("Cannot parse response: %v", err)
		return err
	}

	c.CurrentBalance.Set(b.Balance)
	return nil
}

func (c *Collector) getDadataStats() error {
	t := time.Now().Format("2006-01-02")
	m := fmt.Sprintf("stat/daily?date=%s", t)
	r, err := c.dadataRequest(m)
	if err != nil {
		log.Errorf("Cannot make request for stats: %v", err)
		return err
	}
	defer r.Body.Close()
	s := dadataStats{}

	err = json.NewDecoder(r.Body).Decode(&s)
	if err != nil {
		log.Errorf("Cannot parse response: %v", err)
		return err
	}

	c.ServicesMerging.Set(float64(s.Services.Merging))
	c.ServicesSuggestions.Set(float64(s.Services.Suggestions))
	c.ServicesClean.Set(float64(s.Services.Clean))
	return nil
}

// NewCollector create new collector struct
func NewCollector(url, token, xSecret string) (*Collector, error) {
	c := Collector{}

	if url == "" {
		return nil, fmt.Errorf("URL should not be empty")
	}
	c.dadataAPIURL = url
	if token == "" {
		return nil, fmt.Errorf("Token should not be empty. Please specify it via DADATA_TOKEN env var")
	}
	c.dadataToken = token
	if xSecret == "" {
		return nil, fmt.Errorf("X-Secret should not be empty. Please specify it via DADATA_X_SECRET env var")
	}
	c.dadataXSecret = xSecret

	err := c.dadataCheck()
	if err != nil {
		return nil, err
	}

	c.totalScrapes = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "exporter_scrapes_total",
		Help:      "Count of total scrapes",
	})

	c.failedBalanceScrapes = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "exporter_failed_balance_scrapes_total",
		Help:      "Count of failed balance scrapes",
	})

	c.failedStatsScrapes = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "exporter_failed_stats_scrapes_total",
		Help:      "Count of failed stats scrapes",
	})

	c.CurrentBalance = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "current_balance",
		Help:      "Current balance on Dadata",
	})

	c.ServicesMerging = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "services",
		Name:      "merging_total",
		Help:      "Merging count for today",
	})

	c.ServicesSuggestions = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "services",
		Name:      "suggestions_total",
		Help:      "Suggestions count for today",
	})

	c.ServicesClean = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "services",
		Name:      "clean_total",
		Help:      "Clean count for today",
	})

	return &c, nil
}
