package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func main() {
	valueStr := flag.String("value", "", "Active calories value (float)")
	dateStr := flag.String("date", "", "Date (YYYY-MM-DD), optional")
	filePath := flag.String("file", "", "Target JSONL file path")
	flag.Parse()

	if *valueStr == "" || *filePath == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Input error: missing required args (need --value and --file). Got value=%q date=%q file=%q\n", *valueStr, *dateStr, *filePath)
		os.Exit(2)
	}

	activeCalories, err := strconv.ParseFloat(*valueStr, 64)
	if err != nil || activeCalories < 0 {
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Input error: invalid --value (expected non-negative float). Got value=%q parse_error=%v date=%q file=%q\n", *valueStr, err, *dateStr, *filePath)
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "Input error: invalid --value (expected non-negative float). Got value=%q date=%q file=%q\n", *valueStr, *dateStr, *filePath)
		}
		os.Exit(2)
	}

	entryDate := *dateStr
	if entryDate == "" {
		entryDate = time.Now().Format("2006-01-02")
	} else if _, err := time.Parse("2006-01-02", entryDate); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Input error: invalid --date (expected YYYY-MM-DD). Got date=%q parse_error=%v value=%q file=%q\n", entryDate, err, *valueStr, *filePath)
		os.Exit(2)
	}

	entry := struct {
		Date           string  `json:"date"`
		ActiveCalories float64 `json:"active_calories"`
		Timestamp      int64   `json:"ts"`
	}{
		Date:           entryDate,
		ActiveCalories: activeCalories,
		Timestamp:      time.Now().Unix(),
	}

	b, err := json.Marshal(entry)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Internal error: failed to encode JSON: %v (entry=%+v)\n", err, entry)
		os.Exit(1)
	}

	if err := os.MkdirAll(filepath.Dir(*filePath), 0755); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "IO error: failed to create target directory: %v (file=%q)\n", err, *filePath)
		os.Exit(1)
	}

	f, err := os.OpenFile(*filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "IO error: failed to open target file: %v (file=%q)\n", err, *filePath)
		os.Exit(1)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(string(b) + "\n"); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "IO error: failed to write JSONL: %v (file=%q)\n", err, *filePath)
		os.Exit(1)
	}

	fmt.Println("OK")
}

