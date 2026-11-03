import type { OpenApiConstant } from "@infrastructure/api";

export const ESshCaStatus = {
    /** The one that signs the organization's certificates. */
    Active: "active",
    /** Replaced: it signs no more, and its private key is gone. Listed so servers can be moved off it. */
    Retired: "retired",
} as const;

export type ESshCaStatus = (typeof ESshCaStatus)[keyof typeof ESshCaStatus];

/**
 * An SSH certificate authority, by its public side.
 */
export interface SshCa {
    id: string;
    status: OpenApiConstant<ESshCaStatus>;

    /** The line for sshd's TrustedUserCAKeys file. */
    publicKey: string;

    createdAt: Date;
    retiredAt: Date | null;
}
