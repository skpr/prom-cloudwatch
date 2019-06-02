package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
)

// Interface for interacting with CloudWatch metrics storage.
type Interface interface {
	Add(*cloudwatch.MetricDatum) error
	Flush() error
}

// Client for interacting with CloudWatch metrics storage.
type Client struct {
	logger    Logger
	svc       cloudwatchiface.CloudWatchAPI
	namespace string
	batch     int
	data      []*cloudwatch.MetricDatum
}

// New client for pushing CloudWatch metrics.
func New(logger Logger, svc cloudwatchiface.CloudWatchAPI, namespace string, batch int) (Interface, error) {
	return &Client{
		logger:    logger,
		svc:       svc,
		namespace: namespace,
		batch:     batch,
	}, nil
}

// Add a metric to storage.
func (c *Client) Add(metric *cloudwatch.MetricDatum) error {
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
