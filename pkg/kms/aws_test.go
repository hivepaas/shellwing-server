package kms

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAWSKMSNeedsAKeyAndAWholeCredential(t *testing.T) {
	ctx := context.Background()

	_, err := NewAWSKMS(ctx, AWSKMSOptions{})
	require.ErrorIs(t, err, ErrNotConfigured, "a key has to be named")

	_, err = NewAWSKMS(ctx, AWSKMSOptions{KeyID: "alias/shellwing-dek", AccessKeyID: "AKIA"})
	require.ErrorIs(t, err, ErrNotConfigured, "half a credential is a mistake, not the ambient one")
	assert.Contains(t, err.Error(), "or neither", "and the message says how to ask for the ambient one")

	_, err = NewAWSKMS(ctx, AWSKMSOptions{KeyID: "alias/shellwing-dek", SecretAccessKey: "s"})
	require.ErrorIs(t, err, ErrNotConfigured)
}

// Nothing here reaches AWS: the stored value is refused before a call is made,
// which is what should happen to a backup restored under another provider.
func TestAWSKMSRefusesAKeyAnotherProviderWrapped(t *testing.T) {
	ctx := context.Background()
	wrapper, err := NewAWSKMS(ctx, AWSKMSOptions{
		KeyID: "alias/shellwing-dek", Region: "us-east-1",
		AccessKeyID: "AKIAEXAMPLE", SecretAccessKey: "not-a-real-secret",
	})
	require.NoError(t, err)
	assert.Equal(t, ProviderAWSKMS, wrapper.Name())

	_, err = wrapper.Unwrap(ctx, pack(ProviderVaultTransit, "vault:v1:Zm9v"))
	require.ErrorIs(t, err, ErrWrapped)

	_, err = wrapper.Unwrap(ctx, "hpsalt:c2FsdA== Y2lwaGVy")
	require.ErrorIs(t, err, ErrWrapped, "and one the app secret wrapped")

	_, err = wrapper.Unwrap(ctx, pack(ProviderAWSKMS, "not base64 at all!!"))
	require.ErrorIs(t, err, ErrWrapped, "and one of its own that is not readable")
}
