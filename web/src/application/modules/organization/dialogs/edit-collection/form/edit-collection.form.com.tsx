import React from "react";

import { DialogBody } from "@components/ui/dialog";
import { Field, FieldError, FieldLabel } from "@components/ui/field";
import { Input } from "@components/ui/input";
import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useForm } from "react-hook-form";
import type { Collections_Payload } from "~/organization/api/services/collections-services";

import {
    COLLECTION_DESCRIPTION_MAX_LENGTH,
    COLLECTION_NAME_MAX_LENGTH,
    type EditCollectionFormInput,
    type EditCollectionFormOutput,
    EditCollectionFormSchema,
} from "../schemas";

export function EditCollectionForm({ defaultValues, nameServerError, onSubmit, onHasChanges, footer }: Props) {
    const {
        register,
        handleSubmit,
        formState: { errors, isDirty },
    } = useForm<EditCollectionFormInput, unknown, EditCollectionFormOutput>({
        defaultValues,
        resolver: zodResolver(EditCollectionFormSchema),
        mode: "onSubmit",
    });

    React.useEffect(() => {
        onHasChanges?.(isDirty);
    }, [isDirty, onHasChanges]);

    function onValid(values: EditCollectionFormOutput) {
        onSubmit({
            name: values.name,
            description: values.description,
        });
    }

    function onInvalid(_errors: FieldErrors<EditCollectionFormInput>) {
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
                        htmlFor="collection-name"
                        isRequired
                    >
                        Name
                    </FieldLabel>
                    <Input
                        id="collection-name"
                        {...register("name")}
                        placeholder="Production"
                        maxLength={COLLECTION_NAME_MAX_LENGTH}
                        autoComplete="off"
                        aria-invalid={errors.name !== undefined || nameServerError !== null}
                    />
                    <FieldError
                        errors={[errors.name, nameServerError === null ? undefined : { message: nameServerError }]}
                    />
                </Field>

                <Field>
                    <FieldLabel htmlFor="collection-description">Description</FieldLabel>
                    <Textarea
                        id="collection-description"
                        {...register("description")}
                        rows={2}
                        minRows={2}
                        maxRows={5}
                        maxLength={COLLECTION_DESCRIPTION_MAX_LENGTH}
                        placeholder="Which hosts it holds, and who uses them"
                        aria-invalid={errors.description !== undefined}
                    />
                    <FieldError errors={[errors.description]} />
                </Field>
            </DialogBody>
            {footer}
        </form>
    );
}

interface Props {
    defaultValues: EditCollectionFormInput;
    /** The server's refusal of the name: taken already. */
    nameServerError: string | null;
    onSubmit: (payload: Collections_Payload) => void;
    onHasChanges?: (dirty: boolean) => void;
    footer?: React.ReactNode;
}
