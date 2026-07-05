package schema

import (
	"encoding/json"
	"reflect"
	"testing"
)

type inner struct {
	Known string `json:"known"`
}

type sample struct {
	Name     string          `json:"name"`
	Nested   inner           `json:"nested"`
	Ptr      *inner          `json:"ptr"`
	List     []inner         `json:"list"`
	Raw      json.RawMessage `json:"raw"`
	Ignored  string          `json:"-"`
	unexp    string          //nolint:unused // exercises unexported skip
	Untagged string
}

func TestUnknownFields(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{
			name: "all known",
			raw:  `{"name":"x","nested":{"known":"y"},"untagged":"z"}`,
			want: nil,
		},
		{
			name: "top-level unknown",
			raw:  `{"name":"x","surprise":1}`,
			want: []string{"surprise"},
		},
		{
			name: "nested unknown",
			raw:  `{"nested":{"known":"y","extra":1}}`,
			want: []string{"nested.extra"},
		},
		{
			name: "unknown through pointer field",
			raw:  `{"ptr":{"known":"y","extra":1}}`,
			want: []string{"ptr.extra"},
		},
		{
			name: "unknown inside array element, deduped",
			raw:  `{"list":[{"known":"a","extra":1},{"known":"b","extra":2}]}`,
			want: []string{"list[].extra"},
		},
		{
			name: "raw message contents are opaque",
			raw:  `{"raw":{"anything":{"deeply":"nested"}}}`,
			want: nil,
		},
		{
			name: "json:\"-\" field cannot be matched by its go name",
			raw:  `{"ignored":"present"}`,
			want: []string{"ignored"},
		},
		{
			name: "untagged field matched case-insensitively",
			raw:  `{"UNTAGGED":"z"}`,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnknownFields([]byte(tt.raw), &sample{})
			if err != nil {
				t.Fatalf("UnknownFields: %v", err)
			}
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnknownFields = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnknownFieldsInvalidJSON(t *testing.T) {
	if _, err := UnknownFields([]byte(`{not json`), &sample{}); err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}
