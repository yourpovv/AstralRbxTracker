package settings

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

func readDotEnv(path string) (map[string]string, error) {
	values := map[string]string{}
	file, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return values, nil
	}
	if err != nil {
		return nil, fmt.Errorf("can't read %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		key, value, isPair := strings.Cut(line, "=")
		if !isPair || strings.HasPrefix(line, "#") {
			continue
		}
		values[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"'`)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("can't read %s: %w", path, err)
	}
	return values, nil
}

func SaveWebhook(path, webhookURL string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("can't write to %s: %w", path, err)
	}
	defer file.Close()

	if _, err := fmt.Fprintf(file, "\n%s=%s\n", WebhookKey, webhookURL); err != nil {
		return fmt.Errorf("can't write to %s: %w", path, err)
	}
	return nil
}
