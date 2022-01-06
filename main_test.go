package main

import (
	"testing"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		desc string
		cmd  []byte
		ecmd []byte
		want int
	}{
		{
			desc: "empty",
			want: 0,
		},
		{
			desc: "same",
			cmd:  []byte("kitten"),
			ecmd: []byte("kitten"),
			want: 0,
		},
		{
			desc: "different",
			cmd:  []byte("kitten"),
			ecmd: []byte("sitting"),
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			actual := LevenshteinDistance(tt.cmd, tt.ecmd)
			if tt.want != actual {
				t.Errorf("want %v but got %v\n", tt.want, actual)
			}
		})
	}
}
