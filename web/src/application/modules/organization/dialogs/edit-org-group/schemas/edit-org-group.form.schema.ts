import { z } from "zod";
import { SSH_MAX_PRINCIPALS } from "~/organization/module-shared/constants";
import { createSshPrincipalsTextSchema } from "~/organization/module-shared/schemas";

/** The server's limits on a group. */
export const ORG_GROUP_NAME_MAX_LENGTH = 100;
export const ORG_GROUP_DESCRIPTION_MAX_LENGTH = 500;

/**
 * A group as it is typed: its accounts as one text, split into names by parseSshPrincipals when it is sent.
 */
export const EditOrgGroupFormSchema = z.object({
    name: z
        .string()
        .trim()
        .min(1, "Give the group a name")
        .max(ORG_GROUP_NAME_MAX_LENGTH, `A name can be up to ${ORG_GROUP_NAME_MAX_LENGTH} characters`),
    description: z
        .string()
        .trim()
        .max(
            ORG_GROUP_DESCRIPTION_MAX_LENGTH,
            `A description can be up to ${ORG_GROUP_DESCRIPTION_MAX_LENGTH} characters`,
        ),
    principals: createSshPrincipalsTextSchema(`A group can grant up to ${SSH_MAX_PRINCIPALS} accounts`),
});

export type EditOrgGroupFormInput = z.input<typeof EditOrgGroupFormSchema>;
export type EditOrgGroupFormOutput = z.output<typeof EditOrgGroupFormSchema>;
