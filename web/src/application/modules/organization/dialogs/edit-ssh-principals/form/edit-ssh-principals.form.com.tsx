import React from "react";

import { DialogBody } from "@components/ui/dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { SshPrincipalsField } from "~/organization/module-shared/components";
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

    const { field: principals } = useController({
        control,
        name: "principals",
    });

    React.useEffect(() => {
        onHasChanges?.(isDirty);
    }, [isDirty, onHasChanges]);

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
                <SshPrincipalsField
                    id="ssh-principals"
                    field={principals}
                    error={errors.principals}
                    serverError={serverError}
                    description="Separate them with commas or spaces. Leave it empty to take all of their own away."
                    previewLabel="Their own accounts"
                    emptyText="None of their own"
                />
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
