package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/alecholmes/weathergrid/pkg/model"
)

func newS3Writer(
	bucketName string,
	s3Client *s3.S3,
	logger *log.Logger,
) *s3Writer {
	return &s3Writer{
		bucketName: bucketName,
		s3Client:   s3Client,
		logger:     logger,
	}
}

type s3Writer struct {
	bucketName string
	s3Client   *s3.S3
	logger     *log.Logger
}

var _ snapshotWriter = (*s3Writer)(nil).write

func (s *s3Writer) write(ctx context.Context, snapshot *model.WeatherSnapshot) error {
	objName := fmt.Sprintf("snapshot_%s_default.json", time.Now().Format("20060102150405"))

	blob, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("marshaling snapshot: %w", err)
	}

	s.logger.Printf("putting snapshot object. key=%s", objName)
	_, err = s.s3Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Body:        bytes.NewReader(blob),
		Bucket:      aws.String(s.bucketName),
		ContentType: aws.String("application/json"),
		Key:         aws.String(objName),
	})
	if err != nil {
		return fmt.Errorf("putting snapshot object `%s`: %w", objName, err)
	}

	copySource := fmt.Sprintf("%s/%s", s.bucketName, objName)
	s.logger.Printf("copy snapshot object. source=%s dest=%s", copySource, objName)
	_, err = s.s3Client.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String("snapshot_latest_default.json"),
		ContentType: aws.String("application/json"),
		CopySource:  aws.String(copySource),
		Metadata: map[string]*string{
			"snapshot_object": aws.String(objName),
		},
	})
	if err != nil {
		return fmt.Errorf("copying snapshot object `%s`: %w", objName, err)
	}

	return nil
}
