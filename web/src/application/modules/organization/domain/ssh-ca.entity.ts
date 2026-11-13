import type { OpenApiConstant } from "@infrastructure/api";

export const ESshCaStatus = {
    /** The one that signs the organization's certificates. There is exactly one. */
    Active: "active",
    /**
     * Trusted by the servers, but not signing.
     *
     * Both halves of rotating an authority. A new one is published first, so every server has its key before anything
     * is signed with it - signing straight away would refuse everybody whose server had not been updated. And the one
     * it replaces stays published afterwards, so the certificates it already signed keep working until they expire.
     */
    Published: "published",
    /** Retired: not trusted, not signing, its private key gone, and what it signed revoked. */
    Retired: "retired",
} as const;

export type ESshCaStatus = (typeof ESshCaStatus)[keyof typeof ESshCaStatus];

export const ESshCaKeyProvider = {
    /** The private key is on the server, encrypted. */
    Local: "local",
    /** The private key is in a Vault or OpenBao transit engine, which the server cannot read it out of. */
    VaultTransit: "vault-transit",
} as const;

export type ESshCaKeyProvider = (typeof ESshCaKeyProvider)[keyof typeof ESshCaKeyProvider];

/**
 * An SSH certificate authority, by its public side.
 */
export interface SshCa {
    id: string;
    status: OpenApiConstant<ESshCaStatus>;

    /** The line for sshd's TrustedUserCAKeys file. */
    publicKey: string;

    /** Where its private half is. */
    keyProvider: OpenApiConstant<ESshCaKeyProvider>;

    createdAt: Date;
    retiredAt: Date | null;
}
