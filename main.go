package main

import (
	"fmt"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"

	log "github.com/sirupsen/logrus"
	"github.com/uaraven/cwltail/awsi"
	"github.com/uaraven/cwltail/cwlogs"
	"github.com/uaraven/cwltail/ui"
)

type logCollectionContext struct {
	LogGroup           string
	HighlightPattern   *regexp.Regexp
	LevelDetectPattern *regexp.Regexp
	FilterPattern      *regexp.Regexp
	InvertFilter       bool
	Events             chan cwlogs.CWLEvent
	StartTime          time.Time
	EndTime            *time.Time
}

func createLogLine(context *logCollectionContext, event cwlogs.CWLEvent) *string {
	streamID := event.ShortStreamName()
	var logLine string
	if context.FilterPattern != nil {
		match := context.FilterPattern.FindStringIndex(event.Message())
		if match == nil {
			if !context.InvertFilter {
				return nil
			}
			logLine = event.Message()
		} else {
			if context.InvertFilter {
				return nil
			}
			logLine = ui.HighlightSelection(event.Message(), match, ":cyan")
		}
	} else {
		logLine = event.Message()
	}
	if !options.NoHighlighting {
		if context.HighlightPattern != nil {
			logLine = ui.Colorize(context.HighlightPattern, logLine)
		} else {
			logLine = ui.ColorizeStandard(logLine)
		}
	}
	if options.LevelHighlight {
		if context.LevelDetectPattern != nil {
			matches := context.LevelDetectPattern.FindStringSubmatch(event.Message())
			if len(matches) > 0 {
				logLine = ui.HighlightLogLevel(context.LevelDetectPattern.SubexpNames()[1:], matches[1:], logLine)
			}
		}
	}
	if options.ShowEventTime || options.ShowEventTimestamp {
		var format string
		if options.ShowEventTime {
			format = "15:04:05.000"
		} else {
			format = time.RFC3339
		}
		logLine = fmt.Sprintf("[%s] %s", ui.TimestampColorizer(event.Timestamp().Format(format)), logLine)
	}
	if options.ShowStreamNames {
		logLine = fmt.Sprintf("[%s] %s", ui.StreamNameColorizer(streamID), logLine)
	}
	return &logLine
}

func collectAndDisplay(wg *sync.WaitGroup, context *logCollectionContext) {
	for event := range context.Events {
		logLine := createLogLine(context, event)
		if logLine != nil {
			fmt.Println(*logLine)
		}
	}
	wg.Done()
}

func logTailStream(client *cloudwatchlogs.Client, logGroups []string) {
	logstream := make(chan cwlogs.CWLEvent, 100)
	start := time.Now()

	cwlogs.Log(client, logstream, logGroups, &start, nil)

	logCollectorContext := logCollectionContext{
		LogGroup:  logGroups[0],
		StartTime: start,
		EndTime:   nil,
		Events:    logstream,
	}
	if options.ColorPattern != "" {
		logCollectorContext.HighlightPattern = regexp.MustCompile(options.ColorPattern)
	}

	if options.LevelPattern != "" {
		logCollectorContext.LevelDetectPattern = regexp.MustCompile(options.LevelPattern)
	}
	if options.FilterPattern != "" {
		if options.FilterPattern[0] == '!' {
			logCollectorContext.FilterPattern = regexp.MustCompile(options.FilterPattern[1:])
			logCollectorContext.InvertFilter = true
		} else {
			logCollectorContext.FilterPattern = regexp.MustCompile(options.FilterPattern)
		}

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
	ColorPattern       string   `arg:"-c,--color-pattern" help:"Regex to colorize log lines"`
	ShowStreamNames    bool     `arg:"-s,--show-stream-names" help:"Show shortened stream names"`
	AwsProfile         string   `arg:"-p,--profile" help:"AWS Profile name"`
	AwsDuration        string   `arg:"--duration" help:"AWS Session duration" default:"1h"`
	LevelHighlight     bool     `arg:"-w,--level-highlight" help:"Enable highlighting for log events of WARN and ERROR levels"`
	LevelPattern       string   `arg:"-l,--level-pattern" help:"Regex to extract log level from the log event" default:"(?i)\\b(?:(?P<warning>warn|warning)|(?P<error>error))\\b"`
	DebugLogs          bool     `arg:"--debug-logs" help:"Enable debug logging to debug.log file"`
	FilterPattern      string   `arg:"-f,--filter" help:"Display only lines that match provided regular expression"`
	ShowEventTime      bool     `arg:"-t,--show-event-time" help:"Displays Cloudwatch event time in ISO8601 format. This displays only the time portion of timestamp"`
	ShowEventTimestamp bool     `arg:"-i,--show-event-timestamp" help:"Displays Cloudwatch event timestamp in ISO8601 format"`
	NoHighlighting     bool     `arg:"--no-highlighting" help:"Disables color highlighing of parts of the log message"`
	LogGroups          []string `arg:"positional,required"`
}

func main() {
	arg.MustParse(&options)
	if options.ShowEventTime && options.ShowEventTimestamp {
		fmt.Println("Only one of --show-event-time, --show-event-timestamp options allowed")
		os.Exit(-1)
	}

	if options.DebugLogs {
		logfile, err := os.Create("debug.log")
		if err != nil {
			log.Fatalf("Failed to create log file: %v", err)
		}
		log.SetOutput(logfile)
	} else {
		log.SetLevel(log.WarnLevel)
	}

	duration, err := time.ParseDuration(options.AwsDuration)
	if err != nil {
		fmt.Printf("Failed to parse duration: %v", err)
	}

	client := awsi.CreateCloudwatchLogsClient(awsi.ConfigAWS(options.AwsProfile, duration))

	logTailStream(client, options.LogGroups)
}
