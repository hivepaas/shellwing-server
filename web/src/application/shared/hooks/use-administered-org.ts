import { OrgsQueries } from "@application/shared/data/queries";
import { type Org } from "@application/shared/entities";
import { EOrgRole } from "@application/shared/enums";

interface UseAdministeredOrgResult {
    /** The organization the signed-in user administers; null while loading, or when there is none. */
    org: Org | null;
    isLoading: boolean;
    error: Error | null;
}

/**
 * Whether a role may administer its organization: set up SSH certificates, say who may log in as what.
 */
export function isOrgAdminRole(role: string): boolean {
    return role === EOrgRole.Owner || role === EOrgRole.Admin;
}

/**
 * The organization the signed-in user administers: the first one in which they are an owner or an admin.
 *
 * A self-hosted server has one organization, so this is it, for its admins. The sidebar asks for it on every
 * page, so it is kept for a while; a change to the organization's settings refreshes it.
 */
export function useAdministeredOrg({ enabled = true }: { enabled?: boolean } = {}): UseAdministeredOrgResult {
    const { data, isLoading, error } = OrgsQueries.useFindMany({ staleTime: 5 * 60 * 1000, enabled });
    const org = data?.data.find(item => isOrgAdminRole(item.role)) ?? null;

    return {
        org,
        isLoading,
        error,
    };
}
