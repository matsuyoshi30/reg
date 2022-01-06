package main

import (
	"testing"
)

func TestDamerauLevenshteinDistance(t *testing.T) {
	tests := []struct {
		desc       string
		cmd        []byte
		ecmd       []byte
		w, s, a, d int
		want       int
	}{
		{
			desc: "empty",
			w:    1,
			s:    1,
			a:    1,
			d:    1,
			want: 0,
		},
		{
			desc: "same",
			cmd:  []byte("kitten"),
			ecmd: []byte("kitten"),
			w:    1,
			s:    1,
			a:    1,
			d:    1,
			want: 0,
		},
		{
			desc: "different with balanced weight",
			cmd:  []byte("kitten"),
			ecmd: []byte("sitting"),
			w:    1,
			s:    1,
			a:    1,
			d:    1,
			want: 3,
		},
		{
			desc: "different with unbalanced weight",
			cmd:  []byte("kitten"),
			ecmd: []byte("sitting"),
			w:    0,
			s:    2,
			a:    1,
			d:    3,
			want: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			actual := DamerauLevenshteinDistance(tt.cmd, tt.ecmd, tt.w, tt.s, tt.a, tt.d)
			if tt.want != actual {
				t.Errorf("want %v but got %v\n", tt.want, actual)
			}
		})
	}
}
