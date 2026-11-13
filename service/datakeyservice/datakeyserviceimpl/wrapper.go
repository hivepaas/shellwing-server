package datakeyserviceimpl

import (
	"context"
	"fmt"

	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/kms"
)

// wrapperFor builds the wrapper for one provider, from the configuration.
//
// The provider is a parameter rather than read from the configuration because a
// start has two of them to build: the one configured, which the key is wrapped
// with from now on, and the one that wrapped the key already in the database,
// which is the only one that can open it. They are the same name until an
// operator changes it, and the stored key says which it was - see kms.ProviderOf.
func wrapperFor(ctx context.Context, cfg *config.Config, provider string) (kms.Wrapper, error) {
	switch provider {
	case kms.ProviderLocal:
		wrapper, err := kms.NewLocal(cfg.Secret)
		return wrapper, hperrors.Wrap(err)

	case kms.ProviderVaultTransit:
		wrapper, err := kms.NewVaultTransit(kms.VaultTransitOptions{
			Address:    cfg.KMS.VaultAddress,
			Token:      cfg.KMS.VaultToken,
			Mount:      cfg.KMS.VaultMount,
			Key:        cfg.KMS.VaultKey,
			Namespace:  cfg.KMS.VaultNamespace,
			CACertFile: cfg.KMS.VaultCACertFile,
		})
		return wrapper, hperrors.Wrap(err)

	case kms.ProviderAWSKMS:
		wrapper, err := kms.NewAWSKMS(ctx, kms.AWSKMSOptions{
			KeyID:           cfg.KMS.AWSKeyID,
			Region:          cfg.KMS.AWSRegion,
			AccessKeyID:     cfg.KMS.AWSAccessKeyID,
			SecretAccessKey: cfg.KMS.AWSSecretAccessKey,
			SessionToken:    cfg.KMS.AWSSessionToken,
			EndpointURL:     cfg.KMS.AWSEndpointURL,
		})
		return wrapper, hperrors.Wrap(err)
	}

	// config.KMS.validate refuses an unknown name at load, so reaching this means
	// a provider was added to base.KMSProviders without being added here.
	return nil, hperrors.Wrap(fmt.Errorf("%w: %q", kms.ErrProviderUnknown, provider))
}
