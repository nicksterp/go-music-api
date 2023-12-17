package main

import (
	"bufio"
	"os"
	"strings"
)

func loadEnv(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			continue // comments and empty lines are skipped
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // skips lines that don't match KEY=VAL
		}

		key := parts[0]
		val := parts[1]

		os.Setenv(key, val)
	}

	return scanner.Err()
}
