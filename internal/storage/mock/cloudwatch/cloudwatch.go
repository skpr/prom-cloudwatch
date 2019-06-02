package cloudwatch

import (
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
)

// Client which mocks the CloudFront client.
type Client struct {
	cloudwatchiface.CloudWatchAPI
}

// New mock CloudFront client.
func New() *Client {
	return &Client{}
}

// PutMetricData mock implementation.
func (c *Client) PutMetricData(input *cloudwatch.PutMetricDataInput) (*cloudwatch.PutMetricDataOutput, error) {
	// @todo, Add persistence if required by tests.
	return nil, nil
}
