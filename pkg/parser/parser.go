package parser

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fraser-isbester/git-snap/pkg/types"
)

func ParseEvents(data []byte, format string) ([]types.SnapEvent, error) {
	switch strings.ToLower(format) {
	case "json":
		return parseJSONEvents(data)
	case "jsonl":
		return parseJSONLEvents(data)
	case "csv":
		return parseCSVEvents(data)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func parseJSONEvents(data []byte) ([]types.SnapEvent, error) {
	var events []types.SnapEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return events, nil
}

func parseJSONLEvents(data []byte) ([]types.SnapEvent, error) {
	var events []types.SnapEvent
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var event types.SnapEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return nil, fmt.Errorf("failed to parse JSONL line: %w", err)
		}
		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading JSONL: %w", err)
	}

	return events, nil
}

func parseCSVEvents(data []byte) ([]types.SnapEvent, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	headers := records[0]
	var events []types.SnapEvent

	for i, record := range records[1:] {
		if len(record) != len(headers) {
			return nil, fmt.Errorf("CSV row %d has %d columns, expected %d", i+2, len(record), len(headers))
		}

		event := types.SnapEvent{
			Attributes: make(map[string]interface{}),
			Metadata:   make(map[string]interface{}),
		}

		for j, value := range record {
			header := headers[j]
			switch header {
			case "id":
				event.ID = value
			case "timestamp":
				timestamp, err := parseTimestamp(value)
				if err != nil {
					return nil, fmt.Errorf("invalid timestamp in row %d: %w", i+2, err)
				}
				event.Timestamp = timestamp
			default:
				event.Attributes[header] = parseValue(value)
			}
		}

		events = append(events, event)
	}

	return events, nil
}

func parseTimestamp(value string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}

	for _, format := range formats {
		if timestamp, err := time.Parse(format, value); err == nil {
			return timestamp, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", value)
}

func parseValue(value string) interface{} {
	value = strings.TrimSpace(value)

	if value == "" {
		return ""
	}

	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal
	}

	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	if boolVal, err := strconv.ParseBool(value); err == nil {
		return boolVal
	}

	return value
}

func ParseEventsFromFile(filename string) ([]types.SnapEvent, error) {
	data, err := readFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	format := determineFormat(filename)
	return ParseEvents(data, format)
}

func readFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

func determineFormat(filename string) string {
	if strings.HasSuffix(strings.ToLower(filename), ".json") {
		return "json"
	}
	if strings.HasSuffix(strings.ToLower(filename), ".jsonl") {
		return "jsonl"
	}
	if strings.HasSuffix(strings.ToLower(filename), ".csv") {
		return "csv"
	}
	return "json"
}
