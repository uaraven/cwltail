package awsi

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

// ConfigAWS creates AWS config. If profile is provided it is used as a aws profile name
func ConfigAWS(profile string, sessionDuration time.Duration) *aws.Config {
	var cfg aws.Config
	var err error
	if profile != "" {
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithSharedConfigProfile(profile),
			config.WithAssumeRoleCredentialOptions(func(o *stscreds.AssumeRoleOptions) {
				o.TokenProvider = stscreds.StdinTokenProvider
				o.Duration = sessionDuration
			}))
	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO())

	}

	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	return &cfg
}

// CreateCloudwatchLogsClient creates a clent for Cloudwatch Logs based on provided config
func CreateCloudwatchLogsClient(cfg *aws.Config) *cloudwatchlogs.Client {
	return cloudwatchlogs.NewFromConfig(*cfg)
}
