package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Provider struct {
	client *s3.Client
	bucket string
}

type S3Config struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	Region    string
}

func NewS3Provider(ctx context.Context, cfg S3Config) (*S3Provider, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	})

	p := &S3Provider{client: client, bucket: cfg.Bucket}
	if err := p.EnsureBucket(ctx); err != nil {
		return nil, fmt.Errorf("ensure bucket: %w", err)
	}
	return p, nil
}

func (p *S3Provider) EnsureBucket(ctx context.Context) error {
	_, err := p.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(p.bucket),
	})
	if err == nil {
		return nil
	}
	_, err = p.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(p.bucket),
	})
	return err
}

func (p *S3Provider) Upload(ctx context.Context, key string, reader io.Reader, contentType string, size int64) error {
	_, err := p.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(p.bucket),
		Key:           aws.String(key),
		Body:          reader,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})
	return err
}

func (p *S3Provider) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := p.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (p *S3Provider) Delete(ctx context.Context, key string) error {
	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (p *S3Provider) URL(_ context.Context, key string) (string, error) {
	return fmt.Sprintf("s3://%s/%s", p.bucket, key), nil
}

func (p *S3Provider) Ping(ctx context.Context) error {
	_, err := p.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(p.bucket),
	})
	return err
}

func NewProvider(ctx context.Context, driver, localPath string, s3Cfg S3Config) (Provider, error) {
	switch driver {
	case "s3":
		return NewS3Provider(ctx, s3Cfg)
	default:
		return NewLocalProvider(localPath)
	}
}
