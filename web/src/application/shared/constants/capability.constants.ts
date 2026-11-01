export const CAPABILITY_IDS = {
    SecretReveal: "cap::secret::reveal",
} as const;

export type CapabilityId = (typeof CAPABILITY_IDS)[keyof typeof CAPABILITY_IDS];

export const CAPABILITIES = [
    {
        name: "Can Reveal Secrets",
        id: CAPABILITY_IDS.SecretReveal,
    },
] as const;
