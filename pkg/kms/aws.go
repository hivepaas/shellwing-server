package kms

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awskms "github.com/aws/aws-sdk-go-v2/service/kms"
)

// AWSKMSOptions is what reaching a key in AWS KMS needs.
type AWSKMSOptions struct {
	// KeyID names the key: its ID, its ARN, or an alias such as
	// alias/shellwing-dek. An alias is the one worth using, because rotating to
	// another key is then a change in AWS rather than in this configuration.
	KeyID string
	// Region is the key's. Empty, the ambient configuration decides.
	Region string
	// AccessKeyID and SecretAccessKey are an explicit credential. Left empty, the
	// SDK's own chain is used, which is what picks up an instance role or a
	// service account's - the better way, since there is then no long-lived
	// secret in this server's configuration at all.
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	// EndpointURL overrides where the service is, for an AWS-compatible one or a
	// local test double.
	EndpointURL string
}

// awsKMS wraps the data encryption key with a symmetric key in AWS KMS.
//
// AWS KMS does Encrypt and Decrypt for a symmetric key and never exports it,
// which is all this seam needs. Note for the signer seam later: its asymmetric
// keys do RSA and ECDSA only, with no Ed25519, so an organization whose CA key
// lives in AWS KMS would have to be an ECDSA P-256 one.
type awsKMS struct {
	keyID  string
	client *awskms.Client
}

// NewAWSKMS reaches a key in AWS KMS. It builds the client but does not call it:
// whether the credential works is found on the first wrap or unwrap, at startup.
func NewAWSKMS(ctx context.Context, opts AWSKMSOptions) (Wrapper, error) {
	if opts.KeyID == "" {
		return nil, fmt.Errorf("%w: %s needs a key ID, ARN or alias", ErrNotConfigured, ProviderAWSKMS)
	}

	loaders := []func(*awsconfig.LoadOptions) error{}
	if opts.Region != "" {
		loaders = append(loaders, awsconfig.WithRegion(opts.Region))
	}
	if opts.AccessKeyID != "" || opts.SecretAccessKey != "" {
		if opts.AccessKeyID == "" || opts.SecretAccessKey == "" {
			return nil, fmt.Errorf("%w: %s was given half a credential - set both the access key ID "+
				"and the secret access key, or neither to use the ambient one",
				ErrNotConfigured, ProviderAWSKMS)
		}
		loaders = append(loaders, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				opts.AccessKeyID, opts.SecretAccessKey, opts.SessionToken)))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, loaders...)
	if err != nil {
		return nil, fmt.Errorf("%w: cannot read the AWS configuration: %w", ErrNotConfigured, err)
	}

	options := []func(*awskms.Options){}
	if opts.EndpointURL != "" {
		options = append(options, func(o *awskms.Options) { o.BaseEndpoint = aws.String(opts.EndpointURL) })
	}
	return &awsKMS{keyID: opts.KeyID, client: awskms.NewFromConfig(cfg, options...)}, nil
}

func (w *awsKMS) Name() string { return ProviderAWSKMS }

func (w *awsKMS) Wrap(ctx context.Context, key []byte) (string, error) {
	out, err := w.client.Encrypt(ctx, &awskms.EncryptInput{
		KeyId:     aws.String(w.keyID),
		Plaintext: key,
	})
	if err != nil {
		return "", fmt.Errorf("%s refused to wrap the data encryption key: %w", ProviderAWSKMS, err)
	}
	if len(out.CiphertextBlob) == 0 {
		return "", fmt.Errorf("%w: %s returned none", ErrServiceAnswer, ProviderAWSKMS)
	}
	return pack(ProviderAWSKMS, base64.StdEncoding.EncodeToString(out.CiphertextBlob)), nil
}

func (w *awsKMS) Unwrap(ctx context.Context, stored string) ([]byte, error) {
	payload, err := payloadFor(ProviderAWSKMS, stored)
	if err != nil {
		return nil, err
	}
	blob, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("%w: the stored %s key is not base64: %w", ErrWrapped, ProviderAWSKMS, err)
	}
	// KeyId is given although the blob names its own key: it makes KMS refuse a
	// blob made with a different key rather than decrypt it, which is what turns
	// a restored-from-the-wrong-place backup into an error instead of a mystery.
	out, err := w.client.Decrypt(ctx, &awskms.DecryptInput{
		CiphertextBlob: blob,
		KeyId:          aws.String(w.keyID),
	})
	if err != nil {
		return nil, fmt.Errorf("%s refused to unwrap the data encryption key: %w", ProviderAWSKMS, err)
	}
	return out.Plaintext, nil
}
