package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
)

const apiEndpoint = "https://api.blockpi.io/openapi/v1/rpc"

// RPCRequest represents the request structure for the RPC call
type RPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  []RPCParams `json:"params"`
	ID      int         `json:"id"`
}

// RPCParams represents the parameters in the RPC call
type RPCParams struct {
	APIKey string `json:"apiKey"`
}

// RPCResponse represents the response structure from the RPC call
type RPCResponse struct {
	JSONRPC string     `json:"jsonrpc"`
	ID      int        `json:"id"`
	Result  *RPCResult `json:"result"`
	Error   *RPCError  `json:"error"`
}

// RPCResult represents the result part of the RPC response
type RPCResult struct {
	Balance float64 `json:"balance"`
}
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func getBalance(apiKey string) (float64, error) {
	data := RPCRequest{
		JSONRPC: "2.0",
		Method:  "blockpi_ruBalance",
		Params: []RPCParams{
			{
				APIKey: apiKey,
			},
		},
		ID: 1,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return 0.0, fmt.Errorf("error marshalling request: %w", err)
	}

	resp, err := http.Post(apiEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0.0, fmt.Errorf("error making request: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0.0, fmt.Errorf("error reading response body: %w", err)
	}

	var response RPCResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0.0, fmt.Errorf("error unmarshalling response: %w", err)
	}
	if response.Error != nil {
		return 0.0, fmt.Errorf("error in response: %s", response.Error.Message)
	}
	return response.Result.Balance, nil
}

type Metrics struct {
	up      *prometheus.Desc
	balance *prometheus.Desc
}

func newMetrics() Metrics {
	return Metrics{
		up: prometheus.NewDesc("up",
			"BlockPi is up",
			nil, nil),
		balance: prometheus.NewDesc("account_balance",
			"Balance of BlockPi",
			nil, nil),
	}
}

type Collector struct {
	apiKey  string
	metrics Metrics
}

func newCollector(apiKey string) *Collector {
	return &Collector{
		apiKey:  apiKey,
		metrics: newMetrics(),
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.metrics.balance
	ch <- c.metrics.up
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	balance, err := getBalance(c.apiKey)
	if err != nil {
		// todo: log error
		ch <- prometheus.MustNewConstMetric(c.metrics.up, prometheus.GaugeValue, 0)
		return
	}
	ch <- prometheus.MustNewConstMetric(c.metrics.up, prometheus.GaugeValue, 1)
	ch <- prometheus.MustNewConstMetric(c.metrics.balance, prometheus.GaugeValue, balance)
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.MessageFieldName = "log"
	_ = godotenv.Load()

	apiKey := os.Getenv("BLOCKPI_API_KEY")
	if apiKey == "" {
		log.Fatal().Msg("BLOCKPI_API_KEY is not set")
		panic("BLOCKPI_API_KEY is not set")
	}
	port := os.Getenv("BLOCKPI_LISTEN_PORT")
	if len(port) == 0 {
		port = "8080"
	}

	log.Debug().Msg("Ping BlockPi API")
	_, err := getBalance(apiKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Error when performing initial API testing")
		panic("Error when performing initial API testing")
	}

	collector := newCollector(apiKey)

	prometheus.MustRegister(collector)

	http.Handle("/", http.RedirectHandler("/metrics", http.StatusMovedPermanently))
	http.Handle("/metrics", promhttp.Handler())

	http.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	log.Info().Msgf("Starting BlockPi exporter, listening on port %s", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	log.Info().Msg("Shutting down BlockPi exporter")
}
