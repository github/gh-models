package sse

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type sampleContent struct {
	Name       string `json:"name"`
	NestedData []*struct {
		Count int    `json:"count"`
		Value string `json:"value"`
	} `json:"nested_data"`
}

type badReader struct{}

func (br badReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrClosedPipe
}

func TestEventReader(t *testing.T) {
	t.Run("invalid type", func(t *testing.T) {
		data := []string{
			"invaliddata: {\"name\":\"chatcmpl-7Z4kUpXX6HN85cWY28IXM4EwemLU3\",\"object\":\"chat.completion.chunk\",\"created\":1688594090,\"model\":\"gpt-4-0613\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"content\":\"\"},\"finish_reason\":null}]}\n\n",
		}

		text := strings.NewReader(strings.Join(data, "\n"))
		eventReader := NewEventReader[sampleContent](io.NopCloser(text))

		firstEvent, err := eventReader.Read()
		require.Empty(t, firstEvent)
		require.EqualError(t, err, "unexpected event type: invaliddata")
	})

	t.Run("bad reader", func(t *testing.T) {
		eventReader := NewEventReader[sampleContent](io.NopCloser(badReader{}))
		defer eventReader.Close()

		firstEvent, err := eventReader.Read()
		require.Empty(t, firstEvent)
		require.ErrorIs(t, io.ErrClosedPipe, err)
	})

	t.Run("stream is closed before done", func(t *testing.T) {
		buf := strings.NewReader("data: {}")

		eventReader := NewEventReader[sampleContent](io.NopCloser(buf))

		evt, err := eventReader.Read()
		require.Empty(t, evt)
		require.NoError(t, err)

		evt, err = eventReader.Read()
		require.Empty(t, evt)
		require.EqualError(t, err, "incomplete stream")
	})

	t.Run("spaces around areas", func(t *testing.T) {
		buf := strings.NewReader(
			// spaces between data
			"data: {\"name\":\"chatcmpl-7Z4kUpXX6HN85cWY28IXM4EwemLU3\",\"nested_data\":[{\"count\":0,\"value\":\"with-spaces\"}]}\n" +
				// no spaces
				"data:{\"name\":\"chatcmpl-7Z4kUpXX6HN85cWY28IXM4EwemLU3\",\"nested_data\":[{\"count\":0,\"value\":\"without-spaces\"}]}\n",
		)

		eventReader := NewEventReader[sampleContent](io.NopCloser(buf))

		evt, err := eventReader.Read()
		require.NoError(t, err)
		require.Equal(t, "with-spaces", evt.NestedData[0].Value)

		evt, err = eventReader.Read()
		require.NoError(t, err)
		require.NotEmpty(t, evt)
		require.Equal(t, "without-spaces", evt.NestedData[0].Value)
	})
}
