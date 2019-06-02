package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/rs/xid"
	"github.com/skpr/prometheus-cloudwatch/internal/storage"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	cliPort      = kingpin.Flag("port", "Port which to receive requests.").Envar("PROM_CLOUDWATCH_PORT").Default("8080").Int()
	cliNamespace = kingpin.Flag("namespace", "CloudWatch naemspace to store metrics.").Envar("PROM_CLOUDWATCH_NAMESPACE").Default("prometheus").String()
	cliBatch     = kingpin.Flag("batch", "Number of records to push in a batch.").Envar("PROM_CLOUDWATCH_BATCH").Default("10").Int()
	cliFrequency = kingpin.Flag("frequency", "How frequently to allow a push to CloudWatch.").Envar("PROM_CLOUDWATCH_FREQUENCY").Default("1m").Duration()
	cliVerbose   = kingpin.Flag("verbose", "Print addition debug information.").Envar("PROM_CLOUDWATCH_VERBOSE").Bool()
)

func main() {
	kingpin.Parse()

	svc := cloudwatch.New(session.New())

	lock := time.Now()

	http.HandleFunc("/write", func(w http.ResponseWriter, r *http.Request) {
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

	addr := fmt.Sprintf(":%d", *cliPort)

	log.Infof("Starting server: %s", addr)

	log.Fatal(http.ListenAndServe(addr, nil))
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
