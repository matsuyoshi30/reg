package main

import (
	"bytes"
	"fmt"
	"os"
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
	fmt.Println(string(lcmd))

	// TODO: check command args[0] = git
	// TODO: check command args[1] exists

	// TODO: get LevenshteinDistance between args[1] and git-commands

	// TODO: get candidate command best similarity and execute

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
