import { z } from "zod";
import { SSH_MAX_PRINCIPALS, SSH_PRINCIPAL_PATTERN } from "~/organization/module-shared/constants";
import { parseSshPrincipals } from "~/organization/module-shared/utils";

/**
 * Accounts as they are typed: one text, split into names by parseSshPrincipals when it is sent. The server's rules,
 * for a member's own accounts and a group's alike.
 *
 * @param tooManyMessage What is said when there are more than SSH_MAX_PRINCIPALS.
 */
export function createSshPrincipalsTextSchema(tooManyMessage: string): z.ZodEffects<z.ZodString, string, string> {
    return z.string().superRefine((text, ctx) => {
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
                message: tooManyMessage,
            });
        }
    });
}
