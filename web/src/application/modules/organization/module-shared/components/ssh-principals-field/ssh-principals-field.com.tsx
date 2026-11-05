import type React from "react";

import { Field, FieldDescription, FieldError, FieldLabel } from "@components/ui/field";
import { Textarea } from "@components/ui/textarea";
import { parseSshPrincipals } from "~/organization/module-shared/utils";

import { SshPrincipalBadges } from "../ssh-principal-badges";

/**
 * Accounts typed as one text - separated by commas or spaces - with the names it makes shown below as chips.
 */
export function SshPrincipalsField({ id, field, error, serverError, description, previewLabel, emptyText }: Props) {
    const preview = parseSshPrincipals(field.value);

    return (
        <>
            <Field>
                <FieldLabel htmlFor={id}>Accounts</FieldLabel>
                <Textarea
                    id={id}
                    {...field}
                    rows={3}
                    minRows={3}
                    maxRows={8}
                    placeholder="deploy, ubuntu"
                    className="font-mono"
                    spellCheck={false}
                    autoCapitalize="off"
                    autoCorrect="off"
                    aria-invalid={error !== undefined || serverError !== null}
                />
                <FieldError errors={[error, serverError === null ? undefined : { message: serverError }]} />
                <FieldDescription>{description}</FieldDescription>
            </Field>

            <div className="flex flex-col gap-2">
                <span className="text-sm font-medium">{previewLabel}</span>
                <SshPrincipalBadges
                    principals={preview}
                    emptyText={emptyText}
                />
            </div>
        </>
    );
}

interface Props {
    id: string;
    /** The text field, as react-hook-form's controller gives it. */
    field: React.ComponentProps<typeof Textarea> & { value: string };
    /** The form's own check of the names. */
    error?: { message?: string };
    /** The server's refusal of a name. */
    serverError: string | null;
    description: React.ReactNode;
    /** Above the chips, such as "They may log in as". */
    previewLabel: string;
    /** In place of the chips when no name is typed. */
    emptyText: string;
}
