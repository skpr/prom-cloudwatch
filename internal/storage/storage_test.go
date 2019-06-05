package storage

import (
	"testing"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"

	mockcloudwatch "github.com/skpr/prometheus-cloudwatch/internal/storage/mock/cloudwatch"
	mocklog "github.com/skpr/prometheus-cloudwatch/internal/storage/mock/log"
)

func TestStorage(t *testing.T) {
	var (
		logger    = mocklog.New()
		svc       = mockcloudwatch.New()
		namespace = "test"
		batch     = 2
		whitelist = Whitelist{
			Metrics: []string{
				"metric1",
				"metric2",
				"metric3",
				"metric4",
				"metric6",
			},
			Labels: []string{
				"foo",
			},
		}
	)

	client, err := New(logger, svc, namespace, batch, whitelist)
	assert.Nil(t, err)

	metrics := []prompb.TimeSeries{
		{
			Labels: []prompb.Label{
				{
					Name:  model.MetricNameLabel,
					Value: "metric1",
				},
				{
					Name:  "foo",
					Value: "bar",
				},
			},
			Samples: []prompb.Sample{
				{
					Value: 1,
				},
			},
		},
		{
			Labels: []prompb.Label{
				{
					Name:  model.MetricNameLabel,
					Value: "metric2",
				},
				{
					Name:  "foo",
					Value: "bar",
				},
			},
			Samples: []prompb.Sample{
				{
					Value: 2,
				},
			},
		},
		{
			Labels: []prompb.Label{
				{
					Name:  model.MetricNameLabel,
					Value: "metric3",
				},
				{
					Name:  "foo",
					Value: "bar",
				},
			},
			Samples: []prompb.Sample{
				{
					Value: 3,
				},
			},
		},
		{
			Labels: []prompb.Label{
				{
					Name:  model.MetricNameLabel,
					Value: "metric4",
				},
			},
			Samples: []prompb.Sample{
				{
					Value: 4,
				},
			},
		},
		{
			Labels: []prompb.Label{
				{
					Name:  model.MetricNameLabel,
					Value: "metric5",
				},
			},
			Samples: []prompb.Sample{
				{
					Value: 5,
				},
			},
		},
		{
			Labels: []prompb.Label{
				{
					Name:  model.MetricNameLabel,
					Value: "metric6",
				},
				{
					Name:  "foo",
					Value: "bar",
				},
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
		"Pushing metrics: 2",
		"Skipping because no dimensions were found: metric4",
		"Skipping because metric has not been whitelisted: metric5",
		"Skipping because no values were found: metric6",
		"Pushing metrics: 1",
	}

	assert.Equal(t, logs, logger.Messages)
}
