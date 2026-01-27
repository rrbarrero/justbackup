package domain

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var sizeRegex = regexp.MustCompile(`^([\d.]+)\s*([a-zA-Z]*)$`)

// ParseSize converts a size string like "1.2GB", "500MB", "1024" to bytes.
func ParseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "" {
		return 0, nil
	}

	match := sizeRegex.FindStringSubmatch(sizeStr)
	if match == nil {
		// Try to parse as raw bytes if no unit
		val, err := strconv.ParseInt(sizeStr, 10, 64)
		if err == nil {
			return val, nil
		}
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}

	value, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size value: %s", match[1])
	}

	unit := strings.ToUpper(match[2])
	var multiplier int64 = 1

	switch unit {
	case "B", "":
		multiplier = 1
	case "K", "KB":
		multiplier = 1024
	case "M", "MB":
		multiplier = 1024 * 1024
	case "G", "GB":
		multiplier = 1024 * 1024 * 1024
	case "T", "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("unknown size unit: %s", unit)
	}

	return int64(value * float64(multiplier)), nil
}

// FormatSize converts bytes to a human-readable size string.
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
