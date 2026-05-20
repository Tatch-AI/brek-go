package awssecret

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type fakeSecretsClient struct {
	output *secretsmanager.GetSecretValueOutput
	err    error
	input  *secretsmanager.GetSecretValueInput
}

type stringerValue string

func (s stringerValue) String() string { return string(s) }

func (f *fakeSecretsClient) GetSecretValue(ctx context.Context, input *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	f.input = input
	return f.output, f.err
}

func TestLoaderSuccessAndErrors(t *testing.T) {
	oldLoadConfig := loadDefaultConfig
	oldNewClient := newSecretsManagerClient
	t.Cleanup(func() {
		loadDefaultConfig = oldLoadConfig
		newSecretsManagerClient = oldNewClient
	})

	loadDefaultConfig = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		lo := config.LoadOptions{}
		for _, fn := range optFns {
			if err := fn(&lo); err != nil {
				return aws.Config{}, err
			}
		}
		return aws.Config{Region: lo.Region}, nil
	}

	fake := &fakeSecretsClient{
		output: &secretsmanager.GetSecretValueOutput{
			SecretString: aws.String("secret_value"),
		},
	}
	newSecretsManagerClient = func(cfg aws.Config) secretsClient {
		return fake
	}

	got, err := Loader(map[string]any{
		"key":    "demo",
		"region": "us-west-2",
	})
	if err != nil {
		t.Fatalf("Loader() error = %v", err)
	}
	if got != "secret_value" {
		t.Fatalf("Loader() = %q, want secret_value", got)
	}
	if fake.input == nil || aws.ToString(fake.input.SecretId) != "demo" {
		t.Fatalf("unexpected secret id: %#v", fake.input)
	}

	fake.output = &secretsmanager.GetSecretValueOutput{
		SecretBinary: []byte("binary_value"),
	}
	got, err = Loader(&Params{Key: "demo", Region: "us-west-2"})
	if err != nil {
		t.Fatalf("Loader() binary error = %v", err)
	}
	if got != "YmluYXJ5X3ZhbHVl" {
		t.Fatalf("Loader() binary = %q, want base64", got)
	}

	fake.output = &secretsmanager.GetSecretValueOutput{}
	if _, err := Loader(Params{Key: "demo", Region: "us-west-2"}); err == nil {
		t.Fatal("expected empty value error")
	}

	if _, err := Loader(map[string]any{"key": "demo"}); err == nil {
		t.Fatal("expected missing region error")
	}
	if _, err := Loader(map[string]any{"region": "us-west-2"}); err == nil {
		t.Fatal("expected missing key error")
	}
	if _, err := Loader(map[string]any{"key": 123, "region": 456}); err == nil {
		t.Fatal("expected invalid map value error")
	}
	if _, err := Loader((*Params)(nil)); err == nil {
		t.Fatal("expected nil params pointer error")
	}
	if _, err := Loader("bad"); err == nil {
		t.Fatal("expected invalid param error")
	}

	fake.err = errors.New("boom")
	if _, err := Loader(Params{Key: "demo", Region: "us-west-2"}); err == nil {
		t.Fatal("expected client error")
	}

	loadDefaultConfig = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, errors.New("config error")
	}
	if _, err := Loader(map[string]any{
		"key":    stringerValue("demo"),
		"region": stringerValue("us-west-2"),
	}); err == nil {
		t.Fatal("expected load config error")
	}
}
