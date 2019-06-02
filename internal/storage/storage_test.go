package storage

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/stretchr/testify/assert"

	mockcloudwatch "github.com/skpr/prometheus-cloudwatch/internal/storage/mock/cloudwatch"
	mocklog "github.com/skpr/prometheus-cloudwatch/internal/storage/mock/log"
)

func TestStorage(t *testing.T) {
	var (
		logger    = mocklog.New()
		svc       = mockcloudwatch.New()
		namespace = "test"
		batch     = 10
	)

	client, err := New(logger, svc, namespace, batch)
	assert.Nil(t, err)

	metrics := []*cloudwatch.MetricDatum{
		{
			MetricName: aws.String("metric1"),
			Dimensions: []*cloudwatch.Dimension{},
			Values: []*float64{
				aws.Float64(1),
			},
		},
		{
			MetricName: aws.String("metric2"),
			Dimensions: []*cloudwatch.Dimension{},
			Values: []*float64{
				aws.Float64(2),
			},
		},
		{
			MetricName: aws.String("metric3"),
			Dimensions: []*cloudwatch.Dimension{},
			Values: []*float64{
				aws.Float64(3),
			},
		},
		{
			MetricName: aws.String("metric4"),
			Dimensions: []*cloudwatch.Dimension{},
			Values: []*float64{
				aws.Float64(4),
			},
		},
		{
			MetricName: aws.String("metric5"),
			Dimensions: []*cloudwatch.Dimension{},
			Values: []*float64{
				aws.Float64(5),
			},
		},
		{
			MetricName: aws.String("metric6"),
			Dimensions: []*cloudwatch.Dimension{},
			Values: []*float64{
				aws.Float64(6),
			},
		},
		{
			MetricName: aws.String("metric7"),
			Dimensions: []*cloudwatch.Dimension{},
			Values: []*float64{
				aws.Float64(7),
			},
		},
		{
			MetricName: aws.String("metric8"),
			Dimensions: []*cloudwatch.Dimension{},
			Values: []*float64{
				aws.Float64(8),
			},
		},
		{
			MetricName: aws.String("metric9"),
			Dimensions: []*cloudwatch.Dimension{},
			Values: []*float64{
				aws.Float64(9),
			},
		},
		{
			MetricName: aws.String("metric10"),
			Dimensions: []*cloudwatch.Dimension{},
			Values: []*float64{
				aws.Float64(10),
			},
		},
		{
			MetricName: aws.String("metric11"),
			Dimensions: []*cloudwatch.Dimension{},
			Values: []*float64{
				aws.Float64(11),
			},
		},
	}

	for _, metric := range metrics {
		err = client.Add(metric)
		assert.Nil(t, err)
	}

	err = client.Flush()
	assert.Nil(t, err)

	logs := []string{
		"Pushing metrics: 10",
		"Pushing metrics: 1",
	}

	assert.Equal(t, logs, logger.Messages)
}
