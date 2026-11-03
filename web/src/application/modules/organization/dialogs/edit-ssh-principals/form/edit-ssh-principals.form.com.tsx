import React from "react";

import { Badge } from "@components/ui/badge";
import { DialogBody } from "@components/ui/dialog";
import { Field, FieldDescription, FieldError, FieldLabel } from "@components/ui/field";
import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { parseSshPrincipals } from "~/organization/module-shared/utils";

import {
    type EditSshPrincipalsFormInput,
    type EditSshPrincipalsFormOutput,
    EditSshPrincipalsFormSchema,
} from "../schemas";

export function EditSshPrincipalsForm({ defaultValues, serverError, onSubmit, onHasChanges, footer }: Props) {
    const {
        control,
        handleSubmit,
        formState: { errors, isDirty },
    } = useForm<EditSshPrincipalsFormInput, unknown, EditSshPrincipalsFormOutput>({
        defaultValues,
        resolver: zodResolver(EditSshPrincipalsFormSchema),
        mode: "onSubmit",
    });

    const {
        field: principals,
        fieldState: { invalid: isPrincipalsInvalid },
    } = useController({
        control,
        name: "principals",
    });

    React.useEffect(() => {
        onHasChanges?.(isDirty);
    }, [isDirty, onHasChanges]);

    const preview = parseSshPrincipals(principals.value);

    function onValid(values: EditSshPrincipalsFormOutput) {
        onSubmit(parseSshPrincipals(values.principals));
    }

    function onInvalid(_errors: FieldErrors<EditSshPrincipalsFormInput>) {
        // The errors are shown beside the field.
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                void handleSubmit(onValid, onInvalid)(event);
            }}
            className="min-h-0 flex flex-1 flex-col"
        >
            <DialogBody className="flex flex-col gap-4">
                <Field>
                    <FieldLabel htmlFor="ssh-principals">Accounts</FieldLabel>
                    <Textarea
                        id="ssh-principals"
                        {...principals}
                        rows={3}
                        minRows={3}
                        maxRows={8}
                        placeholder="deploy, ubuntu"
                        className="font-mono"
                        spellCheck={false}
                        autoCapitalize="off"
                        autoCorrect="off"
                        aria-invalid={isPrincipalsInvalid || serverError !== null}
                    />
                    <FieldError
                        errors={[errors.principals, serverError === null ? undefined : { message: serverError }]}
                    />
                    <FieldDescription>
                        Separate them with commas or spaces. Leave it empty to take every account away.
                    </FieldDescription>
                </Field>

                <div className="flex flex-col gap-2">
                    <span className="text-sm font-medium">They may log in as</span>
                    {preview.length > 0 ? (
                        <div className="flex flex-wrap gap-1.5">
                            {preview.map(name => (
                                <Badge
                                    key={name}
                                    tone="neutral"
                                    className="font-mono"
                                >
                                    {name}
                                </Badge>
                            ))}
                        </div>
                    ) : (
                        <span className="text-sm text-muted-foreground">No account: they cannot get a certificate</span>
                    )}
                </div>
            </DialogBody>
            {footer}
        </form>
    );
}

interface Props {
    defaultValues: EditSshPrincipalsFormInput;
    /** The server's refusal of a name, shown beside the field. */
    serverError: string | null;
    /** The names typed, split at commas and spaces, without repeats. */
    onSubmit: (principals: string[]) => void;
    onHasChanges?: (dirty: boolean) => void;
    footer?: React.ReactNode;
}
