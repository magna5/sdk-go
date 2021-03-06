package event_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/event"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
)

func TestTransportSend(t *testing.T) {
	messageChannel := make(chan binding.Message, 1)
	transport := event.NewTransportAdapter(binding.ChanSender(messageChannel), binding.ChanReceiver(messageChannel))
	ev := test.MinEvent()

	client, err := cloudevents.NewClient(transport)
	require.NoError(t, err)

	_, _, err = client.Send(context.Background(), ev)
	require.NoError(t, err)

	result := <-messageChannel

	test.AssertEventEquals(t, ev, cloudevents.Event(result.(event.EventMessage)))
}

func TestTransportReceive(t *testing.T) {
	messageChannel := make(chan binding.Message, 1)
	eventReceivedChannel := make(chan cloudevents.Event, 1)
	transport := event.NewTransportAdapter(binding.ChanSender(messageChannel), binding.ChanReceiver(messageChannel))
	ev := test.MinEvent()

	client, err := cloudevents.NewClient(transport)
	require.NoError(t, err)

	messageChannel <- event.EventMessage(ev)

	go func() {
		err = client.StartReceiver(context.Background(), func(event cloudevents.Event) {
			eventReceivedChannel <- event
		})
		require.NoError(t, err)
	}()

	result := <-eventReceivedChannel

	test.AssertEventEquals(t, ev, result)
}
