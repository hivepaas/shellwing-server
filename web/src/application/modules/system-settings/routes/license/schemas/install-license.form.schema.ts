import { z } from "zod";

/** The server's bound on the text taken as a key: a key is a few kilobytes. */
const MAX_KEY_LENGTH = 32 << 10;

export const InstallLicenseFormSchema = z.object({
    key: z.string().trim().min(1, "Paste a license key").max(MAX_KEY_LENGTH, "This is too long to be a license key"),
});

export type InstallLicenseFormInput = z.input<typeof InstallLicenseFormSchema>;
export type InstallLicenseFormOutput = z.output<typeof InstallLicenseFormSchema>;
