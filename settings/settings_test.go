package settings

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestDuration(t *testing.T) {
	testCases := []struct {
		str   string
		value Duration
	}{
		{`"1s"`, Duration(time.Second)},
		{`"1m"`, Duration(time.Minute)},
		{`"1h"`, Duration(time.Hour)},
		{`null`, 0},
		{`""`, 0},
	}
	codecs := []struct {
		name      string
		marshal   func(interface{}) ([]byte, error)
		unmarshal func([]byte, interface{}) error
	}{
		{"json", json.Marshal, json.Unmarshal},
		{"yaml", yaml.Marshal, yaml.Unmarshal},
	}
	for _, tc := range testCases {
		for _, codec := range codecs {
			t.Run(codec.name, func(t *testing.T) {
				// str --> dur --> mid_str(may different from str) --> dur
				var dur Duration
				err := codec.unmarshal([]byte(tc.str), &dur)
				require.NoError(t, err)
				require.Equal(t, tc.value, dur)

				midStr, err := codec.marshal(dur)
				require.NoError(t, err)
				err = codec.unmarshal(midStr, &dur)
				require.NoError(t, err)
				require.Equal(t, tc.value, dur)
			})
		}
	}
}
