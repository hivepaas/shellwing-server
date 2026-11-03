import { z } from "zod";
import { SSH_MAX_PRINCIPALS, SSH_PRINCIPAL_PATTERN } from "~/organization/module-shared/constants";
import { parseSshPrincipals } from "~/organization/module-shared/utils";

/**
 * The accounts as they are typed: one text, split into names by parseSshPrincipals when it is sent.
 */
export const EditSshPrincipalsFormSchema = z.object({
    principals: z.string().superRefine((text, ctx) => {
        const names = parseSshPrincipals(text);
        const invalid = names.find(name => !SSH_PRINCIPAL_PATTERN.test(name));

        if (invalid !== undefined) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: `"${invalid}" cannot be an account name: use letters, digits and _ . @ -, up to 64, not starting with . @ or -`,
            });

            return;
        }

        if (names.length > SSH_MAX_PRINCIPALS) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: `A member can be granted up to ${SSH_MAX_PRINCIPALS} accounts`,
            });
        }
    }),
});

export type EditSshPrincipalsFormInput = z.input<typeof EditSshPrincipalsFormSchema>;
export type EditSshPrincipalsFormOutput = z.output<typeof EditSshPrincipalsFormSchema>;
