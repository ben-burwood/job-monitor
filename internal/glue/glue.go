package glue

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

const (
	DefaultRegion = "eu-west-2"

	// ObserveRoleArnEnvVar is an optional role to assume for read access to Glue
	ObserveRoleArnEnvVar = "AWS_GLUE_OBSERVE_ROLE_ARN"

	roleSessionName = "data-dashboard"
)

type Client struct {
	region string
	glue   *glue.Client
}

// NewClientFromEnv creates a Glue Client using the default AWS credential chain,
// assuming AWS_GLUE_OBSERVE_ROLE_ARN (auto-refreshing) when set
func NewClientFromEnv(ctx context.Context) (*Client, error) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = DefaultRegion
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	if roleArn := os.Getenv(ObserveRoleArnEnvVar); roleArn != "" {
		provider := stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg), roleArn, func(o *stscreds.AssumeRoleOptions) {
			o.RoleSessionName = roleSessionName
			o.Duration = time.Hour
		})
		cfg.Credentials = aws.NewCredentialsCache(provider)
	}

	return &Client{region: region, glue: glue.NewFromConfig(cfg)}, nil
}

// GetLatestRun returns the most recent run of the job, or nil if it has never run
func (c *Client) GetLatestRun(ctx context.Context, jobName string) (*types.JobRun, error) {
	resp, err := c.glue.GetJobRuns(ctx, &glue.GetJobRunsInput{JobName: aws.String(jobName), MaxResults: aws.Int32(1)})
	if err != nil {
		return nil, err
	}
	if len(resp.JobRuns) == 0 {
		return nil, nil
	}
	return &resp.JobRuns[0], nil
}

// ConsoleURL returns the Glue Studio console link for a job run
func (c *Client) ConsoleURL(jobName, runID string) string {
	return fmt.Sprintf("https://%[1]s.console.aws.amazon.com/gluestudio/home?region=%[1]s#/job/%[2]s/run/%[3]s", c.region, jobName, runID)
}
