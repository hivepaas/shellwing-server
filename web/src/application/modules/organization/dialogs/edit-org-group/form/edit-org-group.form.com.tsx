import React from "react";

import { DialogBody } from "@components/ui/dialog";
import { Field, FieldError, FieldLabel } from "@components/ui/field";
import { Input } from "@components/ui/input";
import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import type { OrgGroups_Payload } from "~/organization/api/services/org-groups-services";
import { SshPrincipalsField } from "~/organization/module-shared/components";
import { parseSshPrincipals } from "~/organization/module-shared/utils";

import {
    type EditOrgGroupFormInput,
    type EditOrgGroupFormOutput,
    EditOrgGroupFormSchema,
    ORG_GROUP_DESCRIPTION_MAX_LENGTH,
    ORG_GROUP_NAME_MAX_LENGTH,
} from "../schemas";

export function EditOrgGroupForm({
    defaultValues,
    nameServerError,
    principalsServerError,
    onSubmit,
    onHasChanges,
    footer,
}: Props) {
    const {
        control,
        register,
        handleSubmit,
        formState: { errors, isDirty },
    } = useForm<EditOrgGroupFormInput, unknown, EditOrgGroupFormOutput>({
        defaultValues,
        resolver: zodResolver(EditOrgGroupFormSchema),
        mode: "onSubmit",
    });

    const { field: principals } = useController({
        control,
        name: "principals",
    });

    React.useEffect(() => {
        onHasChanges?.(isDirty);
    }, [isDirty, onHasChanges]);

    function onValid(values: EditOrgGroupFormOutput) {
        onSubmit({
            name: values.name,
            description: values.description,
            sshPrincipals: parseSshPrincipals(values.principals),
        });
    }

    function onInvalid(_errors: FieldErrors<EditOrgGroupFormInput>) {
        // The errors are shown beside the fields.
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
                    <FieldLabel
                        htmlFor="org-group-name"
                        isRequired
                    >
                        Name
                    </FieldLabel>
                    <Input
                        id="org-group-name"
                        {...register("name")}
                        placeholder="Ops"
                        maxLength={ORG_GROUP_NAME_MAX_LENGTH}
                        autoComplete="off"
                        aria-invalid={errors.name !== undefined || nameServerError !== null}
                    />
                    <FieldError
                        errors={[errors.name, nameServerError === null ? undefined : { message: nameServerError }]}
                    />
                </Field>

                <Field>
                    <FieldLabel htmlFor="org-group-description">Description</FieldLabel>
                    <Textarea
                        id="org-group-description"
                        {...register("description")}
                        rows={2}
                        minRows={2}
                        maxRows={5}
                        maxLength={ORG_GROUP_DESCRIPTION_MAX_LENGTH}
                        placeholder="Who is in it, and why"
                        aria-invalid={errors.description !== undefined}
                    />
                    <FieldError errors={[errors.description]} />
                </Field>

                <SshPrincipalsField
                    id="org-group-ssh-principals"
                    field={principals}
                    error={errors.principals}
                    serverError={principalsServerError}
                    description="Separate them with commas or spaces. Each member may log in as these, on top of their own."
                    previewLabel="Its members may log in as"
                    emptyText="None"
                />
            </DialogBody>
            {footer}
        </form>
    );
}

interface Props {
    defaultValues: EditOrgGroupFormInput;
    /** The server's refusal of the name: taken already. */
    nameServerError: string | null;
    /** The server's refusal of an account name. */
    principalsServerError: string | null;
    onSubmit: (payload: OrgGroups_Payload) => void;
    onHasChanges?: (dirty: boolean) => void;
    footer?: React.ReactNode;
}
