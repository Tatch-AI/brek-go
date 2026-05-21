package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Tatch-AI/brek-go"
)

func TestMain(t *testing.T) {
	oldGetConfig := getConfig
	oldPrintf := printf
	defer func() {
		getConfig = oldGetConfig
		printf = oldPrintf
	}()

	var printed string
	getConfig = func() (map[string]any, error) {
		return map[string]any{"service": "billing"}, nil
	}
	printf = func(format string, args ...any) (int, error) {
		printed = fmt.Sprintf(format, args...)
		return len(printed), nil
	}

	brek.SetLoaders(brek.DefaultLoaders())
	main()

	if !strings.Contains(printed, "service=billing") {
		t.Fatalf("printed output = %q", printed)
	}

	getConfig = func() (map[string]any, error) {
		return nil, os.ErrNotExist
	}
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on config error")
		}
	}()
	main()
}
