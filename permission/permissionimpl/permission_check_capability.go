package permissionimpl

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/permission"
)

// capabilityAction is the only action a capability is ever checked for.
//
// A capability names one operation, so "may execute it" is the whole question.
// Setting it here rather than at every call site removes the one thing a caller
// could get wrong about it - and a caller that asked for the wrong action would
// be refused for a reason that has nothing to do with what they hold.
const capabilityAction = base.ActionTypeExecute

func (p *manager) checkCapability(
	ctx context.Context,
	db database.IDB,
	check *permission.CapabilityCheck,
) (bool, error) {
	normalizeCapabilityCheck(check)

	return p.checkFlatResourceAccess(ctx, db, &check.BaseAccessCheck,
		base.ResourceTypeCapability, string(check.Capability))
}

// normalizeCapabilityCheck forces the action the check is answered on.
//
// It overwrites rather than defaults: leaving a caller's value in place would let
// the check answer a different question than its type promises. Clearing AllOf
// and AnyOf as well keeps exactly one of the three set, which is what BaseAccessCheck.IsValid
// requires.
func normalizeCapabilityCheck(check *permission.CapabilityCheck) {
	check.Action, check.AllOf, check.AnyOf = capabilityAction, nil, nil
}

// HasCapability reports whether the caller holds a capability.
//
// It answers only that: the caller picks the error a refusal turns into, so each
// capability can be refused with its own message even though the lookup behind
// them is the same.
func (p *manager) HasCapability(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	capability base.ResourceCapability,
) (bool, error) {
	hasPerm, err := p.CheckAccess(ctx, db, auth, &permission.CapabilityCheck{
		Capability: capability,
	})
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return hasPerm, nil
}
