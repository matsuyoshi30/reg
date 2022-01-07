package main

import (
	"bytes"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/manifoldco/promptui"
)

func main() {
	var p bool
	flag.BoolVar(&p, "p", false, "Use prompt when multiple command candidates")
	flag.Parse()

	err := run(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error happened: %v\n", err)
		os.Exit(1)
	}
}

func run(prompt bool) error {
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
		candidates[i] = candidate{cmd: cmd}
		if strings.HasPrefix(cmd, string(lcmds[1])) {
			candidates[i].len = 0
		} else {
			candidates[i].len = DamerauLevenshteinDistance(lcmds[1], []byte(cmd), 0, 2, 1, 3)
		}
	}

	sort.Slice(candidates, func(i, j int) bool { return candidates[i].len < candidates[j].len })

	var n int
	for i, candidate := range candidates {
		if candidate.len != 0 {
			break
		}
		n = i + 1
	}

	bestSimilarity := candidates[n].len
	for i := n; i < len(candidates); i++ {
		if candidates[i].len != bestSimilarity {
			break
		}
		n++
	}

	// https://github.com/git/git/blob/dcc0cd074f0c639a0df20461a301af6d45bd582e/help.c#L538-L539
	if bestSimilarity >= 7 {
		return fmt.Errorf("no candidate")
	}

	bests := make([]string, n)
	for i := 0; i < n; i++ {
		bests[i] = candidates[i].cmd
	}

	var execcmd string
	if len(bests) == 1 || !prompt {
		execcmd = bests[0]
	} else {
		prompt := promptui.Select{
			Label: "Select git command you want to execute",
			Items: bests,
		}
		_, result, err := prompt.Run()
		if err != nil {
			return fmt.Errorf("promptui failed: %w", err)
		}
		execcmd = result
	}

	out, err := exec.Command("git", execcmd).Output()
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
		return ReadZshHistory(llb)
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
func ReadZshHistory(line []byte) ([]byte, error) {
	if line[0] != ':' {
		return line, nil
	}

	var ret []byte
	i := bytes.IndexByte(line, ';')
	if i == -1 && i == len(line)-1 {
		return nil, fmt.Errorf("invalid zsh history format")
	}
	ret = line[i+1:]

	return ret, nil
}

type candidate struct {
	cmd string
	len int
}

// DamerauLevenshteinDistance calculates Damerau-Levenshtein distance between cmd and ecmd.
// This implementation allows the costs to be weighted like original git command.
// - w (as in "sWap")
// - s (as in "Substitution")
// - a (for insertion, AKA "Add")
// - d (as in "Deletion")
// ref: https://github.com/git/git/blob/dcc0cd074f0c639a0df20461a301af6d45bd582e/help.c#L606
func DamerauLevenshteinDistance(cmd, ecmd []byte, w, s, a, d int) int {
	l1 := len(cmd)
	l2 := len(ecmd)

	dist := make([][]int, l1+1)
	for i := 0; i <= l1; i++ {
		dist[i] = make([]int, l2+1)

		dist[i][0] = i * d
	}
	for j := 0; j <= l2; j++ {
		dist[0][j] = j * a
	}

	for i := 1; i <= l1; i++ {
		for j := 1; j <= l2; j++ {
			cost := s
			if cmd[i-1] == ecmd[j-1] {
				cost = 0
			}
			dist[i][j] = dist[i-1][j-1] + cost // substitution

			if i > 1 && j > 1 && cmd[i-2] == ecmd[j-1] && cmd[i-1] == ecmd[j-2] {
				dist[i][j] = min(dist[i][j], dist[i-2][j-2]+w) // swap
			}

			dist[i][j] = min(dist[i][j], min(dist[i][j-1]+a, dist[i-1][j]+d)) // add and deletion
		}
	}

	// for debug
	// for i := 0; i <= l1; i++ {
	// 	for j := 0; j < l2; j++ {
	// 		print(dist[i][j])
	// 		print(" ")
	// 	}
	// 	print(dist[i][l2])
	// 	println()
	// }

	return dist[l1][l2]
}

func min(a, b int) int {
	return int(math.Min(float64(a), float64(b)))
}
