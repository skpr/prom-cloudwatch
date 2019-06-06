package utils

import (
	"math"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
)

// TimeSeriesToCloudWatch converts a Prometheus TimeSeries to a CloudWatch MetricDatum.
func TimeSeriesToCloudWatch(ts prompb.TimeSeries, dimensions []string) (*cloudwatch.MetricDatum, error) {
	metric := &cloudwatch.MetricDatum{}

	for _, label := range ts.Labels {
		if label.Name == model.MetricNameLabel {
			metric.MetricName = aws.String(label.Value)
			continue
		}

		if Contains(dimensions, label.Name) {
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

// Contains a string within a slice.
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}
