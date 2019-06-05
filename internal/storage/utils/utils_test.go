package utils

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"
)

func TestTimeSeriesToCloudWatch(t *testing.T) {
	ts := prompb.TimeSeries{
		Labels: []prompb.Label{
			{
				Name:  model.MetricNameLabel,
				Value: "test",
			},
			{
				Name:  "namespace",
				Value: "test",
			},
			{
				Name:  "pod",
				Value: "test",
			},
			{
				Name:  "container",
				Value: "test",
			},
			{
				Name:  "skip",
				Value: "foo",
			},
		},
		Samples: []prompb.Sample{
			{
				Value: 1,
			},
			{
				Value: 2,
			},
			{
				Value: 3,
			},
		},
	}

	labels := []string{
		"namespace",
		"pod",
		"container",
	}

	metric, err := TimeSeriesToCloudWatch(ts, labels)
	assert.Nil(t, err)

	want := &cloudwatch.MetricDatum{
		MetricName: aws.String("test"),
		Dimensions: []*cloudwatch.Dimension{
			{
				Name:  aws.String("namespace"),
				Value: aws.String("test"),
			},
			{
				Name:  aws.String("pod"),
				Value: aws.String("test"),
			},
			{
				Name:  aws.String("container"),
				Value: aws.String("test"),
			},
		},
		Values: []*float64{
			aws.Float64(1),
			aws.Float64(2),
			aws.Float64(3),
		},
	}

	assert.Equal(t, want, metric)
}
