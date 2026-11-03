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
}
