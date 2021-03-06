package event

import (
	"bytes"
	"io"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
)

// Translates a Message with a valid Structured, Binary or Event representation to an Event
// Returns:
// * event, true, false, nil if message was structured and correctly translated to Event
// * event, false, true, nil if message was binary and correctly translated to Event
// * event, false, false, nil if message was event and correctly translated to Event
// * nil, true, false, err if message was structured but error happened during translation
// * nil, false, true, err if message was binary but error happened during translation
// * nil, false, false, err if message was event but error happened during translation
// * nil, false, false, err in other cases
func ToEvent(message binding.Message, factories ...binding.TransformerFactory) (e ce.Event, wasStructured bool, wasBinary bool, err error) {
	e = cloudevents.NewEvent()
	encoder := &messageToEventBuilder{event: &e}
	wasStructured, wasBinary, err = binding.Translate(
		message,
		func() binding.StructuredEncoder {
			return encoder
		},
		func() binding.BinaryEncoder {
			return encoder
		},
		func() binding.EventEncoder {
			return encoder
		},
		factories,
	)
	return e, wasStructured, wasBinary, err
}

type messageToEventBuilder struct {
	event *ce.Event
}

var _ binding.StructuredEncoder = (*messageToEventBuilder)(nil)
var _ binding.BinaryEncoder = (*messageToEventBuilder)(nil)
var _ binding.EventEncoder = (*messageToEventBuilder)(nil)

func (b *messageToEventBuilder) SetEvent(e ce.Event) error {
	b.event.Data = e.Data
	b.event.Context = e.Context.Clone()
	b.event.DataBinary = e.DataBinary
	b.event.DataEncoded = e.DataEncoded
	return nil
}

func (b *messageToEventBuilder) SetStructuredEvent(format format.Format, event io.Reader) error {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, event)
	if err != nil {
		return err
	}
	return format.Unmarshal(buf.Bytes(), b.event)
}

func (b *messageToEventBuilder) SetData(data io.Reader) error {
	var buf bytes.Buffer
	w, err := io.Copy(&buf, data)
	if err != nil {
		return err
	}
	if w != 0 {
		return b.event.SetData(buf.Bytes())
	}
	return nil
}

func (b *messageToEventBuilder) SetAttribute(attribute spec.Attribute, value interface{}) error {
	// If spec version we need to change to right context struct
	if attribute.Kind() == spec.SpecVersion {
		str, err := types.ToString(value)
		if err != nil {
			return err
		}
		switch str {
		case cloudevents.VersionV01:
			b.event.Context = b.event.Context.AsV01()
		case cloudevents.VersionV02:
			b.event.Context = b.event.Context.AsV02()
		case cloudevents.VersionV03:
			b.event.Context = b.event.Context.AsV03()
		case cloudevents.VersionV1:
			b.event.Context = b.event.Context.AsV1()
		}
		return nil
	}
	return attribute.Set(b.event.Context, value)
}

func (b *messageToEventBuilder) SetExtension(name string, value interface{}) error {
	value, err := types.Validate(value)
	if err != nil {
		return err
	}
	b.event.SetExtension(name, value)
	return nil
}
