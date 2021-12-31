package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"

	"github.com/alecholmes/weathergrid/pkg/model"
	"github.com/alecholmes/weathergrid/pkg/openweathermap"
)

func main() {
	local := flag.Bool("local", false, "if true, run locally instead of as lambda")
	configPath := flag.String("config", "", "local config filename")
	snapshotToS3 := flag.Bool("snapshot-to-s3", false, "if true, write snapshots to S3")

	flag.Parse()

	if !*local {
		lambda.Start(HandleRequest)
		return
	}

	cfgData, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("reading config: %v", err)
	}

	log.Printf("using config: %s", string(cfgData))

	var cfg config
	if err := json.Unmarshal(cfgData, &cfg); err != nil {
		log.Fatalf("unmarshaling config: %v", err)
	}

	var writer snapshotWriter
	if *snapshotToS3 {
		sess, err := session.NewSession()
		if err != nil {
			log.Fatalf("creating session: %v", err)
		}

		writer = newS3Writer(cfg.SnapshotsBucketName, s3.New(sess), log.Default()).write
	} else {
		writer = func(_ context.Context, blob *model.WeatherSnapshot) error {
			o, _ := json.MarshalIndent(blob, "", "  ")
			fmt.Println(string(o))
			return nil
		}
	}

	if err := run(context.Background(), &cfg, writer); err != nil {
		log.Fatal(err)
	}
}

type Input struct {
	ConfigBucketName string `json:"config_bucket_name"`
	ConfigObjectName string `json:"config_object_name"`
}

func HandleRequest(ctx context.Context, input Input) (string, error) {
	if input.ConfigBucketName == "" {
		return "", errors.New("config_bucket_name required")
	}
	if input.ConfigObjectName == "" {
		return "", errors.New("config_object_name required")
	}

	sess, err := session.NewSession()
	if err != nil {
		return "", errors.Wrap(err, "creating session")
	}

	cfg, err := loadConfig(ctx, s3.New(sess), input)
	if err != nil {
		return "", fmt.Errorf("loading config: %w", err)
	}

	writer := newS3Writer(cfg.SnapshotsBucketName, s3.New(sess), log.Default())
	if err := run(ctx, cfg, writer.write); err != nil {
		return "", err
	}

	return "ok", nil
}

func run(ctx context.Context, cfg *config, writer snapshotWriter) error {
	weatherClient := openweathermap.NewRealClient(cfg.OpenWeatherMapAPIKey)
	syncer := NewWeatherSyncer(weatherClient, writer, log.Default())
	if err := syncer.Sync(ctx, cfg.Groups); err != nil {
		return errors.Wrap(err, "syncing weather")
	}

	return nil
}

func loadConfig(ctx context.Context, s3Client *s3.S3, input Input) (*config, error) {
	configResp, err := s3Client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(input.ConfigBucketName),
		Key:    aws.String(input.ConfigObjectName),
	})
	if err != nil {
		return nil, fmt.Errorf("fetching config: %w", err)
	}
	defer configResp.Body.Close()

	var cfg config
	if err := json.NewDecoder(configResp.Body).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	return &cfg, nil
}
