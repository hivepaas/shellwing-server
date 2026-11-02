package blobstore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"

	"github.com/hivepaas/shellwing-server/hperrors"
)

// S3Config is where blobs go in S3-compatible storage. Prefix is a folder in the bucket, or empty. PathStyle is for
// servers that have no bucket subdomains, as MinIO usually has not.
type S3Config struct {
	Endpoint  string
	Region    string
	Bucket    string
	Prefix    string
	AccessKey string
	SecretKey string
	PathStyle bool
}

type s3Store struct {
	client *s3.Client
	bucket string
	prefix string
}

// NewS3 keeps blobs in a bucket. Plain PUTs are all it needs: conditional writes are decided in the database, so any
// S3 look-alike will do.
func NewS3(cfg *S3Config) (Store, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("%w: the bucket is not set", ErrKeyInvalid)
	}
	client := s3.New(s3.Options{
		Region:       cfg.Region,
		UsePathStyle: cfg.PathStyle,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		// Look-alikes differ on trailing checksums; sending them only where asked is what all of them take.
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	}, func(opts *s3.Options) {
		if cfg.Endpoint != "" {
			opts.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	})
	prefix := strings.Trim(cfg.Prefix, "/")
	if prefix != "" {
		prefix += "/"
	}
	return &s3Store{client: client, bucket: cfg.Bucket, prefix: prefix}, nil
}

func (s *s3Store) Check(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(s.bucket)})
	return hperrors.Wrap(err)
}

func (s *s3Store) objectKey(key string) (*string, error) {
	if !validKey(key) {
		return nil, fmt.Errorf("%w: %q", ErrKeyInvalid, key)
	}
	return aws.String(s.prefix + key), nil
}

func (s *s3Store) Put(ctx context.Context, key string, data []byte) error {
	objectKey, err := s.objectKey(key)
	if err != nil {
		return err
	}
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    objectKey,
		Body:   bytes.NewReader(data),
	})
	return hperrors.Wrap(err)
}

func (s *s3Store) Get(ctx context.Context, key string) ([]byte, error) {
	objectKey, err := s.objectKey(key)
	if err != nil {
		return nil, err
	}
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: objectKey})
	if err != nil {
		if isS3NotFound(err) {
			return nil, ErrNotFound
		}
		return nil, hperrors.Wrap(err)
	}
	defer func() { _ = out.Body.Close() }()
	data, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return data, nil
}

func (s *s3Store) Delete(ctx context.Context, key string) error {
	objectKey, err := s.objectKey(key)
	if err != nil {
		return err
	}
	_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: objectKey})
	if err != nil && !isS3NotFound(err) {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *s3Store) List(ctx context.Context, prefix string, fn func(Blob) error) error {
	pages := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(s.prefix + prefix),
	})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return hperrors.Wrap(err)
		}
		for i := range page.Contents {
			object := &page.Contents[i]
			err = fn(Blob{
				Key:      strings.TrimPrefix(aws.ToString(object.Key), s.prefix),
				Size:     aws.ToInt64(object.Size),
				Modified: aws.ToTime(object.LastModified),
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func isS3NotFound(err error) bool {
	var noSuchKey *types.NoSuchKey
	if errors.As(err, &noSuchKey) {
		return true
	}
	var notFound *types.NotFound
	if errors.As(err, &notFound) {
		return true
	}
	var apiErr smithy.APIError
	return errors.As(err, &apiErr) && (apiErr.ErrorCode() == "NoSuchKey" || apiErr.ErrorCode() == "NotFound")
}
