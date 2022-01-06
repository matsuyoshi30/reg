package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error happened: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// get command executed the most recently
	lcmd, err := getCommandExecutedJustBefore()
	if err != nil {
		return err
	}

	lcmds := bytes.Split(lcmd, []byte(" "))
	if !bytes.Equal([]byte("git"), lcmds[0]) {
		return fmt.Errorf("not git command executed just before")
	}
	if len(lcmds) < 2 {
		return fmt.Errorf("no git command")
	}

	candidates := make([]candidate, len(commands))
	for i, cmd := range commands {
		candidates[i] = candidate{cmd: cmd, len: levenshteinDistance([]byte(cmd), lcmds[1])}
	}

	bestSimilarity := candidates[0].len
	bests := make([]candidate, 0)
	for _, candidate := range candidates {
		if bestSimilarity > candidate.len {
			bestSimilarity = candidate.len
		}
	}

	// https://github.com/git/git/blob/dcc0cd074f0c639a0df20461a301af6d45bd582e/help.c#L538-L539
	if bestSimilarity > 7 {
		return nil
	}
	for _, candidate := range candidates {
		if candidate.len == bestSimilarity {
			bests = append(bests, candidate)
		}
	}

	// TODO: select command when multiple candidates
	out, err := exec.Command("git", bests[0].cmd).Output()
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}
	fmt.Println(string(out))

	return nil
}

func getCommandExecutedJustBefore() ([]byte, error) {
	// detect current working shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		return nil, fmt.Errorf("unable to detect current working shell")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	if shell == "/bin/zsh" {
		llb, err := readHistFile(home)
		if err != nil {
			return nil, fmt.Errorf("failed to read history file last line: %w", err)
		}
		return formatZshHistory(llb)
	}

	return nil, fmt.Errorf("does not support yet")
}

func readHistFile(home string) (ll []byte, err error) {
	f, err := os.Open(filepath.Join(home, ".zsh_history"))
	if err != nil {
		return nil, fmt.Errorf("failed to get history file: %w", err)
	}
	defer func() {
		err = f.Close()
	}()

	stat, err := os.Stat(filepath.Join(home, ".zsh_history"))
	if err != nil {
		return nil, fmt.Errorf("failed to get history file stat: %w", err)
	}

	buf := make([]byte, 256)
	start := stat.Size() - int64(len(buf))
	_, err = f.ReadAt(buf, start)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var isLast int
	for i := len(buf) - 2; i >= 0; i-- {
		if buf[i] == '\n' {
			if isLast == 0 {
				isLast = i
				continue
			}
			ll = buf[i+1 : isLast]
			break
		}
	}

	return ll, nil
}

// https://stackoverflow.com/questions/37961165/how-zsh-stores-history-history-file-format
func formatZshHistory(line []byte) ([]byte, error) {
	if line[0] != ':' {
		return line, nil
	}

	var ret []byte
	if i := bytes.IndexByte(line, ';'); i == -1 {
		return nil, fmt.Errorf("invalid zsh history format")
	} else {
		ret = line[i+1:]
		if len(ret) == 0 {
			return nil, fmt.Errorf("invalid zsh history format")
		}
	}

	return ret, nil
}

type candidate struct {
	cmd string
	len int
}

// TODO: weight
// https://github.com/git/git/blob/dcc0cd074f0c639a0df20461a301af6d45bd582e/help.c#L606
func levenshteinDistance(cmd, ecmd []byte) int {
	l1 := len(cmd)
	l2 := len(ecmd)

	dist := make([][]int, l1+1)
	for i := 0; i <= l1; i++ {
		dist[i] = make([]int, l2+1)

		dist[i][0] = i
	}
	for j := 0; j <= l2; j++ {
		dist[0][j] = j
	}

	for i := 1; i <= l1; i++ {
		for j := 1; j <= l2; j++ {
			cost := 1
			if cmd[i-1] == ecmd[j-1] {
				cost = 0
			}
			dist[i][j] = min(dist[i-1][j-1]+cost, min(dist[i][j-1]+1, dist[i-1][j]+1))
		}
	}

	return dist[l1][l2]
}

func min(a, b int) int {
	return int(math.Min(float64(a), float64(b)))
}
