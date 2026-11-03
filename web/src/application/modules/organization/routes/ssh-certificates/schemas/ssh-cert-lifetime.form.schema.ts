import { z } from "zod";
import { SSH_CERT_TTL_MAX_MINUTES, SSH_CERT_TTL_MIN_MINUTES } from "~/organization/module-shared/constants";

const RANGE_MESSAGE = `From ${SSH_CERT_TTL_MIN_MINUTES} to ${SSH_CERT_TTL_MAX_MINUTES} minutes`;

export const SshCertLifetimeFormSchema = z.object({
    certTtlMinutes: z
        .number({ required_error: "Enter a number of minutes", invalid_type_error: "Enter a number of minutes" })
        .int("A whole number of minutes")
        .min(SSH_CERT_TTL_MIN_MINUTES, RANGE_MESSAGE)
        .max(SSH_CERT_TTL_MAX_MINUTES, RANGE_MESSAGE),
});

export type SshCertLifetimeFormInput = z.input<typeof SshCertLifetimeFormSchema>;
export type SshCertLifetimeFormOutput = z.output<typeof SshCertLifetimeFormSchema>;
