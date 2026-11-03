import type { OpenApiConstant } from "@infrastructure/api";

export const ELicenseStatus = {
    /** No key is installed: the free limit holds. */
    None: "none",
    Valid: "valid",
    Trial: "trial",
    /** Past its expiry, and still in force for its grace. */
    Grace: "grace",
    /** Past its grace: the free limit is back. */
    Expired: "expired",
    /** The installed key no longer verifies. */
    Invalid: "invalid",
    /** The installed key names another server. */
    OtherInstallation: "other-installation",
} as const;

export type ELicenseStatus = (typeof ELicenseStatus)[keyof typeof ELicenseStatus];

/**
 * The license in force on this server, and the seats taken.
 */
export interface License {
    status: OpenApiConstant<ELicenseStatus>;

    /** This server's ID, which a key is issued for. */
    installationId: string;

    /** The key's plan while it holds; empty on the free limit. */
    plan: string;

    /** How many users may hold a seat; null: no limit. */
    maxUsers: number | null;

    /** The limit with no license in force, which is back when one ends. */
    freeUsers: number;

    /** How many hold one now: the active and the invited. */
    usedUsers: number;

    trial: boolean;
    customerName: string;
    expiresAt: Date | null;
    graceEndsAt: Date | null;
}
