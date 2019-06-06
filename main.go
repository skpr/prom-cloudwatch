package main

import (
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/heptio/workgroup"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/prometheus/prompb"
	"github.com/rs/xid"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"

	"github.com/skpr/prometheus-cloudwatch/internal/storage"
)

var (
	cliAddress   = kingpin.Flag("address", "Address which this writer will respond to requests.").Envar("PROMETHUES_CLOUDWATCH_ADDRESS").Default(":8080").String()
	cliNamespace = kingpin.Flag("namespace", "CloudWatch naemspace to store metrics.").Envar("PROMETHUES_CLOUDWATCH_NAMESPACE").Default("prometheus").String()
	cliBatch     = kingpin.Flag("batch", "Number of records to push in a batch.").Envar("PROMETHUES_CLOUDWATCH_BATCH").Default("10").Int()
	cliWhitelist = kingpin.Flag("whitelist", "Path to whitelist configuration file.").Envar("PROMETHUES_CLOUDWATCH_WHITELIST").Required().String()
	cliFrequency = kingpin.Flag("frequency", "How frequently to allow a push to CloudWatch.").Envar("PROMETHUES_CLOUDWATCH_FREQUENCY").Default("1m").Duration()
	cliVerbose   = kingpin.Flag("verbose", "Print addition debug information.").Envar("PROMETHUES_CLOUDWATCH_VERBOSE").Bool()
	cliExporter  = kingpin.Flag("exporter", "Address which Prometheus exporter metrics can be scraped.").Envar("PROMETHUES_CLOUDWATCH_EXPORTER").Default(":9000").String()
)

func main() {
	kingpin.Parse()

	wg := workgroup.Group{}

	// Expose metrics for debugging.
	wg.Add(metrics)

	// Start writing metrics.
	wg.Add(writer)

	if err := wg.Run(); err != nil {
		panic(err)
	}
}

// Starts to Prometheus writer.
func writer(stop <-chan struct{}) error {
	lock := time.Now()

	mux := http.NewServeMux()

	var whitelist storage.Whitelist

	file, err := ioutil.ReadFile(*cliWhitelist)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(file, &whitelist)
	if err != nil {
		return err
	}

	mux.HandleFunc("/write", func(w http.ResponseWriter, r *http.Request) {
		logger := log.With("request", xid.New())

		if time.Now().Before(lock) {
			if *cliVerbose {
				log.Infof("Skipping request will store new requests after: %s", lock.String())
			}

			return
		}

		compressed, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		reqBuf, err := snappy.Decode(nil, compressed)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var req prompb.WriteRequest

		if err := proto.Unmarshal(reqBuf, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		lock = time.Now().Add(*cliFrequency)

		svc := cloudwatch.New(session.New())

		client, err := storage.New(logger, svc, *cliNamespace, *cliBatch, whitelist)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, ts := range req.Timeseries {
			err = client.Add(ts)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		err = client.Flush()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	listen, err := net.Listen("tcp", *cliAddress)
	if err != nil {
		return err
	}

	go func() {
		<-stop
		listen.Close()
	}()

	log.Infof("Starting writer server: %s", *cliAddress)

	return http.Serve(listen, mux)
}

// Exposes Prometheus metrics.
func metrics(stop <-chan struct{}) error {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())

	listen, err := net.Listen("tcp", *cliExporter)
	if err != nil {
		return err
	}

	go func() {
		<-stop
		listen.Close()
	}()

	log.Infof("Starting metrics servere: %s", *cliExporter)

	return http.Serve(listen, mux)
}
