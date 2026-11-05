import { z } from "zod";
import { SSH_MAX_PRINCIPALS } from "~/organization/module-shared/constants";
import { createSshPrincipalsTextSchema } from "~/organization/module-shared/schemas";

/**
 * The member's own accounts as they are typed: one text, split into names by parseSshPrincipals when it is sent.
 */
export const EditSshPrincipalsFormSchema = z.object({
    principals: createSshPrincipalsTextSchema(`A member can be granted up to ${SSH_MAX_PRINCIPALS} accounts`),
});

export type EditSshPrincipalsFormInput = z.input<typeof EditSshPrincipalsFormSchema>;
export type EditSshPrincipalsFormOutput = z.output<typeof EditSshPrincipalsFormSchema>;
