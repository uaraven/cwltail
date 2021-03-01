package cwlogs

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"

	log "github.com/sirupsen/logrus"
)

const (
	renewalDelay = 15 * time.Second
)

type LogStreams interface {
	Get() *logStream
	Update(stream logStream)
}

type logStreamsImpl struct {
	sync.RWMutex
	streams []logStream
}

func (ls *logStreamsImpl) Get() *logStream {
	ls.RLock()
	defer ls.RUnlock()
	if len(ls.streams) > 0 {
		return &ls.streams[0]
	}
	return nil
}

func (ls *logStreamsImpl) Update(logs logStream) {
	ls.Lock()
	defer ls.Unlock()
	var sn []string
	if len(logs.streamNames) > 100 {
		log.Tracef("Too many stream names %d, taking last 100", len(logs.streamNames))
		sn = logs.streamNames[len(logs.streamNames)-100:]
	} else {
		sn = logs.streamNames
	}

	ls.streams = []logStream{
		{
			logGroup:    logs.logGroup,
			streamNames: sn,
		},
	}
}

type LogStreamingContext struct {
	Client       *cloudwatchlogs.Client
	Streams      LogStreams
	Dedupe       Deduplicator
	StartTime    *time.Time
	EndTime      *time.Time
	EventChannel chan CWLEvent
}

type logStream struct {
	logGroup    string
	streamNames []string
}

// CWLEvent contains data that comprises an cloudwatch logs event
type CWLEvent interface {
	EventID() string
	Timestamp() time.Time
	Message() string
	LogGroup() string
	LogStream() string
	ShortStreamName() string
}

type cwlEventImpl struct {
	eventID   string
	timestamp time.Time
	message   string
	logGroup  string
	logStream string
}

func (c cwlEventImpl) EventID() string {
	return c.eventID
}

func (c cwlEventImpl) Timestamp() time.Time {
	return c.timestamp
}

func (c cwlEventImpl) Message() string {
	return strings.TrimRight(c.message, "\n\r")
}

func (c cwlEventImpl) LogGroup() string {
	return c.logGroup
}

func (c cwlEventImpl) LogStream() string {
	return c.logStream
}

func (c cwlEventImpl) ShortStreamName() string {
	return c.logStream[len(c.logStream)-6:]
}

// getStreams returns a slice of logGroup/streamName pairs for each passed log group
func (ctx *LogStreamingContext) getStreams(logGroups []string) ([]logStream, error) {
	result := make([]logStream, 0)

	groupStreams := make(map[string][]string, 0)

	for _, logGroup := range logGroups {
		params := &cloudwatchlogs.DescribeLogStreamsInput{
			LogGroupName: aws.String(logGroup),
			OrderBy:      types.OrderByLastEventTime,
			Descending:   aws.Bool(true),
		}

		paginator := cloudwatchlogs.NewDescribeLogStreamsPaginator(ctx.Client, params)
	out:
		for paginator.HasMorePages() {
			log.Tracef("Next page within log group %s", logGroup)
			output, err := paginator.NextPage(context.TODO())
			if err != nil {
				// TODO: Handle rate error and force timeout
				return nil, err
			}
			for _, s := range output.LogStreams {
				log.Tracef("Stream %s, last event: %d, start time: %d", *s.LogStreamName, *s.LastEventTimestamp, TimeToAws(*ctx.StartTime))
				if ctx.EndTime != nil {
					if *s.LastEventTimestamp < TimeToAws(*ctx.StartTime) ||
						*s.FirstEventTimestamp > TimeToAws(*ctx.EndTime) {
						// for range mode ignore everything that ends before or starts after the range
						break out
					}
				} else if (TimeToAws(*ctx.StartTime) - *s.LastEventTimestamp) > 3600000 {
					// for tailing mode ignore all streams that have last event from more than an hour ago
					break out
				}
				if val, ok := groupStreams[logGroup]; ok {
					if len(val) < 100 {
						groupStreams[logGroup] = append(val, *s.LogStreamName)
					} else {
						log.Tracef("Too many streams for %s, ignoring", logGroup)
					}
				} else {
					v := make([]string, 1)
					v[0] = *s.LogStreamName
					groupStreams[logGroup] = v
				}
			}
			log.Tracef("Total streams: %d", len(groupStreams[logGroup]))
			time.Sleep(100 * time.Millisecond)
		}
	}

	for k, v := range groupStreams {
		log.Tracef("Group: %s, Streams: %d", k, len(v))
		result = append(result, logStream{
			logGroup:    k,
			streamNames: v,
		})
	}

	return result, nil
}

func (ctx *LogStreamingContext) readEventsFromLogGroup() {
	var starting int64
	var ending int64
	if ctx.Dedupe.GetLastTimestamp() == 0 {
		starting = TimeToAws(*ctx.StartTime)
	} else {
		starting = ctx.Dedupe.GetLastTimestamp()
	}

	stream := ctx.Streams.Get()
	if stream == nil {
		log.Traceln("No streams found")
		return
	}
	log.Tracef("Log streams %v", stream.streamNames)

	params := cloudwatchlogs.FilterLogEventsInput{
		LogGroupName:   aws.String(stream.logGroup),
		LogStreamNames: stream.streamNames,
		StartTime:      aws.Int64(starting),
	}
	if ctx.EndTime != nil {
		ending = TimeToAws(*ctx.EndTime)
		params.EndTime = aws.Int64(ending)
	}
	log.Tracef("Get events from group %s, # of streams: %d", stream.logGroup, len(stream.streamNames))

	paginator := cloudwatchlogs.NewFilterLogEventsPaginator(ctx.Client, &params)
	if paginator.HasMorePages() {
		log.Tracef("Reading next page of events from %s", stream.logGroup)
		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			log.Errorln(err)
			// TODO: Send error to err channel
		} else {
			log.Tracef("Got %d events from %s", len(output.Events), stream.logGroup)
		}
		for _, e := range output.Events {
			ctx.Dedupe.AddAndExecuteIfNotPresent(*e.EventId, *e.Timestamp, func() {
				cwlEvent := &cwlEventImpl{
					eventID:   *e.EventId,
					logGroup:  stream.logGroup,
					logStream: *e.LogStreamName,
					timestamp: AwsToTime(*e.Timestamp),
					message:   *e.Message,
				}
				ctx.EventChannel <- cwlEvent
			})
		}
	}
	log.Traceln("Stream read done")
	if ctx.EndTime != nil {
		log.Traceln("Closing event channel")
		close(ctx.EventChannel)
	}
}

func (ctx *LogStreamingContext) streamRenewal(t *time.Ticker, logGroups []string) {
	for range t.C {
		log.Traceln("Stream renewal time")
		streams, err := ctx.getStreams(logGroups[0:1])
		if err == nil && len(streams) > 0 {
			ctx.Streams.Update(streams[0])
		}
	}
}

// Log starts reading events from the client and posting them to eventChannel
// Currently only the first log group from logGroups is used
func Log(client *cloudwatchlogs.Client, eventChannel chan CWLEvent, logGroups []string, startTime *time.Time, endTime *time.Time) {
	ctx := LogStreamingContext{
		Client:       client,
		Dedupe:       NewDeduplicator(-1, -1),
		EventChannel: eventChannel,
		StartTime:    startTime,
		EndTime:      endTime,
		Streams:      &logStreamsImpl{},
	}
	streams, err := ctx.getStreams(logGroups[0:1])
	if err != nil {
		log.Fatalln(err)
	}

	if len(streams) > 0 {
		ctx.Streams.Update(streams[0])
	}

	if ctx.StartTime == nil {
		s := time.Now()
		ctx.StartTime = &s
	}

	if endTime == nil {
		log.Traceln("Tailing CWL")

		// check for new streams every now and then
		t := time.NewTicker(renewalDelay)
		go ctx.streamRenewal(t, logGroups)

		logCheck := time.NewTicker(250 * time.Millisecond)
		go func() {
			for range logCheck.C {
				ctx.readEventsFromLogGroup()
			}
		}()

	} else {
		log.Traceln("Period CWL")

		go func() {
			ctx.readEventsFromLogGroup()
		}()
	}

}
