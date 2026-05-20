package awssecret

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type Params struct {
	Key    string
	Region string
}

type secretsClient interface {
	GetSecretValue(context.Context, *secretsmanager.GetSecretValueInput, ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

var loadDefaultConfig = config.LoadDefaultConfig

var newSecretsManagerClient = func(cfg aws.Config) secretsClient {
	return secretsmanager.NewFromConfig(cfg)
}

func Loader(params any) (string, error) {
	p, err := parseParams(params)
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	cfg, err := loadDefaultConfig(ctx, config.WithRegion(p.Region))
	if err != nil {
		return "", err
	}

	client := newSecretsManagerClient(cfg)
	out, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(p.Key),
	})
	if err != nil {
		return "", err
	}

	switch {
	case out.SecretString != nil:
		return aws.ToString(out.SecretString), nil
	case len(out.SecretBinary) > 0:
		return base64.StdEncoding.EncodeToString(out.SecretBinary), nil
	default:
		return "", fmt.Errorf("awsSecret: secret %q returned empty value", p.Key)
	}
}

func parseParams(params any) (Params, error) {
	switch v := params.(type) {
	case Params:
		return validateParams(v)
	case *Params:
		if v == nil {
			return Params{}, errors.New("awsSecret: params must include key and region")
		}
		return validateParams(*v)
	case map[string]any:
		return validateParams(Params{
			Key:    stringFromMap(v, "key"),
			Region: stringFromMap(v, "region"),
		})
	default:
		return Params{}, errors.New("awsSecret: params must be an object with key and region")
	}
}

func validateParams(p Params) (Params, error) {
	p.Key = strings.TrimSpace(p.Key)
	p.Region = strings.TrimSpace(p.Region)

	if p.Key == "" || p.Region == "" {
		return Params{}, errors.New("awsSecret: params must include key and region")
	}

	return p, nil
}

func stringFromMap(m map[string]any, key string) string {
	value, _ := m[key]
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return ""
	}
}
