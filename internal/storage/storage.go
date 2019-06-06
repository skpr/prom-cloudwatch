package storage

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/prometheus/prometheus/prompb"

	storageutils "github.com/skpr/prometheus-cloudwatch/internal/storage/utils"
)

// Interface for interacting with CloudWatch metrics storage.
type Interface interface {
	Add(prompb.TimeSeries) error
	Flush() error
}

// Client for interacting with CloudWatch metrics storage.
type Client struct {
	logger    Logger
	svc       cloudwatchiface.CloudWatchAPI
	namespace string
	batch     int
	whitelist Whitelist
	data      []*cloudwatch.MetricDatum
}

// Whitelist which governs which metrics are pushed to CloudWatch.
type Whitelist struct {
	Metrics []string `json:"metrics" yaml:"metrics"`
	Labels  []string `json:"labels"  yaml:"labels"`
}

// New client for pushing CloudWatch metrics.
func New(logger Logger, svc cloudwatchiface.CloudWatchAPI, namespace string, batch int, whitelist Whitelist) (Interface, error) {
	client := &Client{
		logger:    logger,
		svc:       svc,
		namespace: namespace,
		batch:     batch,
		whitelist: whitelist,
	}

	if len(whitelist.Metrics) == 0 {
		return client, errors.New("metrics whitelist was not provided")
	}

	if len(whitelist.Labels) == 0 {
		return client, errors.New("labels whitelist was not provided")
	}

	return client, nil
}

// Add a metric to storage.
func (c *Client) Add(ts prompb.TimeSeries) error {
	metric, err := storageutils.TimeSeriesToCloudWatch(ts, c.whitelist.Labels)
	if err != nil {
		return err
	}

	if !storageutils.Contains(c.whitelist.Metrics, *metric.MetricName) {
		c.logger.Infof("Skipping because metric has not been whitelisted: %s", *metric.MetricName)
		return nil
	}

	if len(metric.Dimensions) == 0 {
		c.logger.Infof("Skipping because no dimensions were found: %s", *metric.MetricName)
		return nil
	}

	if len(metric.Values) == 0 {
		c.logger.Infof("Skipping because no values were found: %s", *metric.MetricName)
		return nil
	}

	c.data = append(c.data, metric)

	if len(c.data) >= c.batch {
		return c.Flush()
	}

	return nil
}

// Flush all records kept in memory.
func (c *Client) Flush() error {
	if len(c.data) > 0 {
		c.logger.Infof("Pushing metrics: %d", len(c.data))

		input := &cloudwatch.PutMetricDataInput{
			Namespace:  aws.String(c.namespace),
			MetricData: c.data,
		}

		_, err := c.svc.PutMetricData(input)
		if err != nil {
			return err
		}

		c.data = nil
	}

	return nil
}
