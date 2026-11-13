import type { EOrgRole } from "@application/shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

/**
 * An organization, as the signed-in user, one of its members, sees it.
 */
export interface Org {
    id: string;
    slug: string;
    name: string;

    /** The signed-in user's role in it. */
    role: OpenApiConstant<EOrgRole>;

    /** How long the SSH certificates it signs last, in minutes. */
    sshCertTtlMinutes: number;

    /** The accounts the signed-in user's certificates may log in as. */
    sshPrincipals: string[];

    /** What a certificate it signs is allowed to do, by sshd's extension names. Empty: nothing but logging in. */
    sshCertExtensions: string[];

    /**
     * When its servers were given a token to fetch the revocation list with, or null when they have none. The token
     * itself is shown once, in the answer that makes it, and never again.
     */
    sshKrlTokenCreatedAt: Date | null;
}
