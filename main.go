package main

import (
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/heptio/workgroup"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/rs/xid"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/skpr/prometheus-cloudwatch/internal/storage"
)

var (
	cliAddress   = kingpin.Flag("address", "Address which this writer will respond to requests.").Envar("PROM_CLOUDWATCH_ADDRESS").Default(":8080").String()
	cliNamespace = kingpin.Flag("namespace", "CloudWatch naemspace to store metrics.").Envar("PROM_CLOUDWATCH_NAMESPACE").Default("prometheus").String()
	cliBatch     = kingpin.Flag("batch", "Number of records to push in a batch.").Envar("PROM_CLOUDWATCH_BATCH").Default("10").Int()
	cliFrequency = kingpin.Flag("frequency", "How frequently to allow a push to CloudWatch.").Envar("PROM_CLOUDWATCH_FREQUENCY").Default("1m").Duration()
	cliVerbose   = kingpin.Flag("verbose", "Print addition debug information.").Envar("PROM_CLOUDWATCH_VERBOSE").Bool()
	cliExporter  = kingpin.Flag("exporter", "Address which Prometheus exporter metrics can be scraped.").Envar("PROM_CLOUDWATCH_EXPORTER").Default(":9000").String()
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

		client, err := storage.New(logger, svc, *cliNamespace, *cliBatch)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, ts := range req.Timeseries {
			metric, err := extractMetric(ts)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if len(metric.Values) == 0 {
				continue
			}

			err = client.Add(metric)
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

// Helper function to extract a metric from the time series data.
func extractMetric(ts prompb.TimeSeries) (*cloudwatch.MetricDatum, error) {
	metric := &cloudwatch.MetricDatum{}

	for _, label := range ts.Labels {
		if label.Name == model.MetricNameLabel {
			metric.MetricName = aws.String(label.Value)
		} else {
			metric.Dimensions = append(metric.Dimensions, &cloudwatch.Dimension{
				Name:  aws.String(label.Name),
				Value: aws.String(label.Value),
			})
		}
	}

	for _, sample := range ts.Samples {
		if !math.IsNaN(sample.Value) {
			metric.Values = append(metric.Values, aws.Float64(sample.Value))
		}
	}

	return metric, nil
}
