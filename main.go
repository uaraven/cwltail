package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	log "github.com/sirupsen/logrus"
	"github.com/uaraven/cwltail/cwlogs"
	"github.com/uaraven/cwltail/ui"
)

type logCollectionContext struct {
	LogGroup           string
	HighlightPattern   *regexp.Regexp
	LevelDetectPattern *regexp.Regexp
	Events             chan cwlogs.CWLEvent
	StartTime          time.Time
	EndTime            *time.Time
}

func createLogLine(context *logCollectionContext, event cwlogs.CWLEvent) string {
	var level string
	if context.LevelDetectPattern != nil {
		level = context.LevelDetectPattern.FindString(event.Message())
	}
	streamID := event.ShortStreamName()
	var logLine string
	if context.HighlightPattern == nil {
		logLine = event.Message()
	} else {
		logLine = ui.Colorize(context.HighlightPattern, event.Message())
	}
	if level != "" {
		logLine = ui.HighlightLogLevel(level, logLine)
	}
	if options.ShowStreamNames {
		logLine = fmt.Sprintf("%s %s", ui.StreamName("["+streamID+"]"), logLine)
	}
	return logLine
}

func collectAndDisplay(wg *sync.WaitGroup, context *logCollectionContext) {
	for event := range context.Events {
		fmt.Println(createLogLine(context, event))
	}
	wg.Done()
}

func createCWLClient() *cloudwatchlogs.Client {

	cfg, err := config.LoadDefaultConfig(context.TODO())

	if options.AwsProfile != "" {
		// Create the credentials from AssumeRoleProvider to assume the role
		// referenced by the "myRoleARN" ARN using the MFA token code provided.
		creds := stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg), options.AwsProfile, func(o *stscreds.AssumeRoleOptions) {
			o.TokenProvider = stscreds.StdinTokenProvider
		})

		cfg.Credentials = creds // &aws.CredentialsCache{Provider: creds}

		// cfg, err = config.LoadDefaultConfig(context.TODO(), stscreds.)
	}

	if err != nil {
		panic(err)
	}

	client := cloudwatchlogs.NewFromConfig(cfg)
	return client
}

func logTailStream(client *cloudwatchlogs.Client, logGroups []string) {
	logstream := make(chan cwlogs.CWLEvent, 100)
	start := time.Now()

	cwlogs.Log(client, logstream, logGroups, &start, nil)

	logCollectorContext := logCollectionContext{
		LogGroup:         logGroups[0],
		HighlightPattern: regexp.MustCompile(`(\d{2}:\d{2}:\d{2}.\d{3})\s+\[(.*)\]\s+(\S+)\s+([a-zA-Z0-9_.]+).*`),
		StartTime:        start,
		EndTime:          nil,
		Events:           logstream,
	}

	if options.LevelPattern != "" {
		logCollectorContext.LevelDetectPattern = regexp.MustCompile(options.LevelPattern)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go collectAndDisplay(&wg, &logCollectorContext)

	wg.Wait()
}

type positional struct {
	LogStream string
}

var options struct {
	ColorPattern    string   `arg:"-c,--color-pattern" help:"Regex to colorize log lines"`
	ShowStreamNames bool     `arg:"-s,--show-stream-names" help:"Show shortened stream names"`
	AwsProfile      string   `arg:"-p,--profile" help:"AWS Profile name"`
	LevelPattern    string   `arg:"-l,--level-pattern" help:"Regex to extract log level from the log event"`
	LogGroups       []string `arg:"positional,required"`
}

func main() {

	arg.MustParse(&options)

	log.SetLevel(log.TraceLevel)
	logfile, err := os.Create("log.log")
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logfile)

	client := createCWLClient()

	logTailStream(client, options.LogGroups)
}
