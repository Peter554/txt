package main

import (
	"bufio"
	"os"
	"path/filepath"
)

type point struct {
	x, y int
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func copyRunes(line []rune) []rune {
	o := make([]rune, len(line))
	copy(o, line)
	return o
}

func copyRunes2D(text [][]rune) [][]rune {
	o := make([][]rune, len(text))
	for idx, line := range text {
		o[idx] = copyRunes(line)
	}
	return o
}

func readFile(path string) ([][]rune, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()
	o := make([][]rune, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		o = append(o, []rune(line))
	}
	return o, nil
}

func writeFile(path string, text [][]rune) error {
	if err := os.MkdirAll(filepath.Dir(path), 0733); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, line := range text {
		_, err := file.WriteString(string(line) + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}
