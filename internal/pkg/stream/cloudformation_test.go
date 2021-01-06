package stream

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stretchr/testify/require"
)

type mockCloudFormation struct {
	out *cloudformation.DescribeStackEventsOutput
	err error
}

func (m mockCloudFormation) DescribeStackEvents(*cloudformation.DescribeStackEventsInput) (*cloudformation.DescribeStackEventsOutput, error) {
	return m.out, m.err
}

func TestStackStreamer_Subscribe(t *testing.T) {
	// GIVEN
	streamer := &StackStreamer{}
	sub1 := make(chan StackEvent)
	sub2 := make(chan StackEvent)

	// WHEN
	streamer.Subscribe(sub1, sub2)

	// THEN
	require.Equal(t, 2, len(streamer.subscribers), "expected number of subscribers to match")
	require.ElementsMatch(t, []chan StackEvent{sub1, sub2}, streamer.subscribers)
}

func TestStackStreamer_Fetch(t *testing.T) {
	t.Run("stores all events in chronological order on fetch", testStackStreamer_Fetch_Success)
	t.Run("stores only events after the changeset creation time", testStackStreamer_Fetch_PostChangeSet)
	t.Run("stores only events that have not been seen yet", testStackStreamer_Fetch_WithSeenEvents)
	t.Run("returns wrapped error if describe call fails", testStackStreamer_Fetch_WithError)
}

func TestStackStreamer_Notify(t *testing.T) {
	// GIVEN
	wantedEvents := []StackEvent{
		{
			LogicalResourceID: "Cluster",
			ResourceType:      "AWS::ECS::Cluster",
			ResourceStatus:    "CREATE_COMPLETE",
		},
		{
			LogicalResourceID: "PublicLoadBalancer",
			ResourceType:      "AWS::ElasticLoadBalancingV2::LoadBalancer",
			ResourceStatus:    "CREATE_COMPLETE",
		},
	}
	sub := make(chan StackEvent, 2)
	streamer := &StackStreamer{
		subscribers:   []chan StackEvent{sub},
		eventsToFlush: wantedEvents,
	}

	// WHEN
	streamer.Notify()
	close(sub) // Close the channel to stop expecting to receive new events.

	// THEN
	var actualEvents []StackEvent
	for event := range sub {
		actualEvents = append(actualEvents, event)
	}
	require.ElementsMatch(t, wantedEvents, actualEvents)
}

func testStackStreamer_Fetch_Success(t *testing.T) {
	// GIVEN
	client := mockCloudFormation{
		// Events are in reverse chronological order.
		out: &cloudformation.DescribeStackEventsOutput{
			StackEvents: []*cloudformation.StackEvent{
				{
					EventId:              aws.String("1"),
					LogicalResourceId:    aws.String("CloudformationExecutionRole"),
					ResourceStatus:       aws.String("CREATE_FAILED"),
					ResourceStatusReason: aws.String("phonetool-test-CFNExecutionRole already exists"),
					Timestamp:            aws.Time(time.Date(2020, time.November, 23, 19, 0, 0, 0, time.UTC)),
				},
				{
					EventId:           aws.String("2"),
					LogicalResourceId: aws.String("Cluster"),
					ResourceStatus:    aws.String("CREATE_COMPLETE"),
					Timestamp:         aws.Time(time.Date(2020, time.November, 23, 18, 0, 0, 0, time.UTC)),
				},
				{
					EventId:           aws.String("3"),
					LogicalResourceId: aws.String("PublicLoadBalancer"),
					ResourceStatus:    aws.String("CREATE_COMPLETE"),
					Timestamp:         aws.Time(time.Date(2020, time.November, 23, 17, 0, 0, 0, time.UTC)),
				},
			},
		},
	}
	streamer := NewStackStreamer(client, "phonetool-test", time.Date(2020, time.November, 23, 16, 0, 0, 0, time.UTC))

	// WHEN
	_, err := streamer.Fetch()

	// THEN
	require.NoError(t, err)
	require.ElementsMatch(t, []StackEvent{
		{
			LogicalResourceID:    "CloudformationExecutionRole",
			ResourceStatus:       "CREATE_FAILED",
			ResourceStatusReason: "phonetool-test-CFNExecutionRole already exists",
		},
		{
			LogicalResourceID: "PublicLoadBalancer",
			ResourceStatus:    "CREATE_COMPLETE",
		},
		{
			LogicalResourceID: "Cluster",
			ResourceStatus:    "CREATE_COMPLETE",
		},
	}, streamer.eventsToFlush, "expected eventsToFlush to appear in chronological order")
}

func testStackStreamer_Fetch_PostChangeSet(t *testing.T) {
	// GIVEN
	client := mockCloudFormation{
		out: &cloudformation.DescribeStackEventsOutput{
			StackEvents: []*cloudformation.StackEvent{
				{
					EventId:           aws.String("abc"),
					LogicalResourceId: aws.String("Cluster"),
					ResourceStatus:    aws.String("CREATE_COMPLETE"),
					Timestamp:         aws.Time(time.Date(2020, time.November, 23, 18, 0, 0, 0, time.UTC)),
				},
			},
		},
	}
	streamer := &StackStreamer{
		client:                client,
		stackName:             "phonetool-test",
		changeSetCreationTime: time.Date(2020, time.November, 23, 19, 0, 0, 0, time.UTC), // An hour after the last event.
	}

	// WHEN
	_, err := streamer.Fetch()

	// THEN
	require.NoError(t, err)
	require.Empty(t, streamer.eventsToFlush, "expected eventsToFlush to be empty")
}

func testStackStreamer_Fetch_WithSeenEvents(t *testing.T) {
	// GIVEN
	client := mockCloudFormation{
		out: &cloudformation.DescribeStackEventsOutput{
			StackEvents: []*cloudformation.StackEvent{
				{
					EventId:           aws.String("abc"),
					LogicalResourceId: aws.String("Cluster"),
					ResourceStatus:    aws.String("CREATE_COMPLETE"),
					Timestamp:         aws.Time(time.Date(2020, time.November, 23, 18, 0, 0, 0, time.UTC)),
				},
				{
					EventId:           aws.String("def"),
					LogicalResourceId: aws.String("PublicLoadBalancer"),
					ResourceStatus:    aws.String("CREATE_COMPLETE"),
					Timestamp:         aws.Time(time.Date(2020, time.November, 23, 17, 0, 0, 0, time.UTC)),
				},
			},
		},
	}
	streamer := &StackStreamer{
		client:                client,
		stackName:             "phonetool-test",
		changeSetCreationTime: time.Date(2020, time.November, 23, 16, 0, 0, 0, time.UTC),
		pastEventIDs: map[string]bool{
			"def": true,
		},
	}

	// WHEN
	_, err := streamer.Fetch()

	// THEN
	require.NoError(t, err)
	require.ElementsMatch(t, []StackEvent{
		{
			LogicalResourceID: "Cluster",
			ResourceStatus:    "CREATE_COMPLETE",
		},
	}, streamer.eventsToFlush, "expected only the event not seen yet to be flushed")
}

func testStackStreamer_Fetch_WithError(t *testing.T) {
	// GIVEN
	client := mockCloudFormation{
		err: errors.New("some error"),
	}
	streamer := &StackStreamer{
		client:                client,
		stackName:             "phonetool-test",
		changeSetCreationTime: time.Date(2020, time.November, 23, 16, 0, 0, 0, time.UTC),
	}

	// WHEN
	_, err := streamer.Fetch()

	// THEN
	require.EqualError(t, err, "describe stack events phonetool-test: some error")
}
