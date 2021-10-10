package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"

	"github.com/alecholmes/weathergrid/pkg/model"
)

func main() {
	lambda.Start(HandleRequest)
}

type Input struct {
	BucketName string `json:"bucket_name"`
}

func HandleRequest(ctx context.Context, input Input) (string, error) {
	if input.BucketName == "" {
		return "", errors.New("bucket_name required")
	}

	sess, err := session.NewSession()
	if err != nil {
		return "", errors.Wrap(err, "creating session")
	}

	syncer := NewWeatherSyncer(input.BucketName, s3.New(sess))
	blobKey, err := syncer.Sync(ctx)
	if err != nil {
		return "", errors.Wrap(err, "syncing weather")
	}

	return blobKey, nil
}

func NewWeatherSyncer(
	bucketName string,
	s3Client *s3.S3,
) *WeatherSyncer {
	return &WeatherSyncer{
		bucketName: bucketName,
		nowFunc:    time.Now,
		s3Client:   s3Client,
	}
}

type WeatherSyncer struct {
	bucketName string
	nowFunc    func() time.Time
	s3Client   *s3.S3
}

func (w *WeatherSyncer) Sync(ctx context.Context) (string, error) {
	now := w.nowFunc().UTC()

	data := &model.WeatherBlob{
		Locations: []*model.LocationWeather{
			{
				Name: "Test-SF",
			},
		},
	}

	blob, err := json.Marshal(data)
	if err != nil {
		return "", errors.Wrap(err, "marshaling blob json")
	}

	objKey := fmt.Sprintf("weather_snapshot_%s.json", now.Format(time.RFC3339))

	if _, err := w.s3Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Body:        bytes.NewReader(blob),
		Bucket:      &w.bucketName,
		ContentType: aws.String("application/json"),
		Key:         aws.String(objKey),
		Metadata:    map[string]*string{},
	}); err != nil {
		return "", errors.Wrapf(err, "putting object %s", objKey)
	}

	return objKey, nil
}
