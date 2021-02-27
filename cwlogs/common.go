package cwlogs

import (
	"strings"
	"time"
)

// Arn object
type Arn interface {
	ShortName() string
	Arn() string
}

type arnImpl struct {
	arn string
}

func (a arnImpl) Arn() string {
	return a.arn
}

func (a arnImpl) ShortName() string {
	parts := strings.Split(a.arn, "/")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return a.arn
}

// NewArn creates a new Arn
func NewArn(arn string) Arn {
	return &arnImpl{
		arn: arn,
	}
}

// LogGroupConfig contains configuration parameters of awslogs driver, including group name, stream prefix and aws region
type LogGroupConfig struct {
	Group        string
	StreamPrefix string
	Region       string
}

// TimeToAws converts a Time to a millisecond epoch timestamp
func TimeToAws(tm time.Time) int64 {
	return tm.Unix() * 1000
}

// AwsToTime converts millisecond epoch timestamp to a Time
func AwsToTime(ts int64) time.Time {
	return time.Unix(ts/1000, 0)
}

// AwsToMs converts millisecond epoch timestamp to a seconds epoch time
func AwsToMs(ts int64) int64 {
	return ts / 1000
}
