package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestReadHistFile(t *testing.T) {
	tmpdir := t.TempDir()

	tests := []struct {
		desc     string
		hContent string
		noFile   bool
		want     []byte
		isErr    bool
	}{
		{
			desc: "empty",
		},
		{
			desc: "normal",
			hContent: `git statu
reg
`,
			want: []byte("git statu"),
		},
		{
			desc: "one line history",
			hContent: `reg
`,
		},
		{
			desc: "no file",
			hContent: `git statu
reg
`,
			noFile: true,
			isErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var (
				hf  *os.File
				err error
			)
			if !tt.noFile {
				hf, err = os.Create(filepath.Join(tmpdir, tt.desc))
				if err != nil {
					t.Fatal(err)
				}

				_, err = hf.Write([]byte(tt.hContent))
				if err != nil {
					t.Fatal(err)
				}

				if err := hf.Close(); err != nil {
					t.Fatal(err)
				}
			}

			actual, err := ReadHistFile(tmpdir, tt.desc)
			if tt.isErr {
				if err == nil {
					t.Errorf("want error but got nil\n")
				}
				return
			}
			if err != nil {
				t.Errorf("want no error but got %v\n", err)
				return
			}
			if !bytes.Equal(tt.want, actual) {
				t.Errorf("want '%v' but got '%v'\n", string(tt.want), string(actual))
			}
		})
	}
}

func TestReadZshHistory(t *testing.T) {
	tests := []struct {
		desc  string
		line  []byte
		want  []byte
		isErr bool
	}{
		{
			desc: "empty",
		},
		{
			desc: "normal",
			line: []byte("git status"),
			want: []byte("git status"),
		},
		{
			desc: "EXTENDED_HISTORY format",
			line: []byte(": 1641393282:0;git status"),
			want: []byte("git status"),
		},
		{
			desc:  "invalid (no timestamp value)",
			line:  []byte(": :0;git status"),
			isErr: true,
		},
		{
			desc:  "invalid (no elapsed seconds value)",
			line:  []byte(": 1641393282:;git status"),
			isErr: true,
		},
		{
			desc:  "invalid (no command)",
			line:  []byte(": 1641393282:0;"),
			isErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			actual, err := ReadZshHistory(tt.line)
			if tt.isErr {
				if err == nil {
					t.Errorf("want error but got nil\n")
				}
				return
			}
			if err != nil {
				t.Errorf("want no error but got %v\n", err)
				return
			}
			if !bytes.Equal(tt.want, actual) {
				t.Errorf("want %v but got %v\n", tt.want, actual)
			}
		})
	}
}

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
