package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"go.viam.com/rdk/app"
	"go.viam.com/rdk/logging"
)

func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		os.Setenv(strings.TrimSpace(key), strings.TrimSpace(val))
	}
	return scanner.Err()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: cli <dataset_id>")
		os.Exit(1)
	}
	datasetID := os.Args[1]

	ctx := context.Background()
	logger := logging.NewLogger("cli")

	if err := loadDotEnv(".env"); err != nil && !os.IsNotExist(err) {
		panic(fmt.Sprintf("failed to load .env: %v", err))
	}

	viamClient, err := app.CreateViamClientFromEnvVars(ctx, nil, logger)
	if err != nil {
		panic(fmt.Sprintf("failed to create viam client: %v", err))
	}
	defer viamClient.Close()

	dataClient := viamClient.DataClient()

	resp, err := dataClient.BinaryDataByFilter(ctx, false, &app.DataByFilterOptions{
		Filter: &app.Filter{DatasetID: datasetID},
		Limit:  10,
	})
	if err != nil {
		panic(fmt.Sprintf("BinaryDataByFilter failed: %v", err))
	}

	logger.Infof("count=%d last=%q", len(resp.BinaryData), resp.Last)
	for i, d := range resp.BinaryData {
		if d.Metadata != nil {
			logger.Infof("[%d] BinaryDataID=%s ID=%s", i, d.Metadata.BinaryDataID, d.Metadata.ID)
		}
	}
}
