import { useState } from "react";

import { Button } from "@components/ui/button";
import {
    Dialog,
    DialogActionFooter,
    DialogDescription,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { toast } from "sonner";
import type { Collections_Payload } from "~/organization/api/services/collections-services";
import { CollectionsCommands } from "~/organization/data/commands";
import type { Collection } from "~/organization/domain";

import { isAlreadyExistException } from "@infrastructure/api";

import { EditCollectionForm } from "../form";
import { useEditCollectionDialogState } from "../hooks";

export function EditCollectionDialog() {
    const { state, props: dialogProps, ...actions } = useEditCollectionDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const open = state.mode !== "closed";
    const isEdit = state.mode === "edit";

    function onSuccess(collection: Collection) {
        setHasChanges(false);
        actions.close();
        dialogProps?.onSuccess?.(collection);
    }

    const create = CollectionsCommands.useCreateOne({
        onSuccess: response => {
            toast.success(`Collection ${response.data.name} created. Grant access to it next.`);
            onSuccess(response.data);
        },
    });

    const update = CollectionsCommands.useUpdateOne({
        onSuccess: response => {
            toast.success(`Collection ${response.data.name} saved`);
            onSuccess(response.data);
        },
    });

    const { isPending, error } = isEdit ? update : create;

    const nameServerError =
        error !== null && isAlreadyExistException(error) ? "Another collection has this name already" : null;

    function handleClose() {
        if (isPending) {
            return;
        }

        if (hasChanges) {
            const confirmed = window.confirm("Are you sure you want to close without saving changes?");

            if (!confirmed) {
                return;
            }
        }

        setHasChanges(false);
        create.reset();
        update.reset();
        actions.close();
        dialogProps?.onClose?.();
    }

    function onSubmit(payload: Collections_Payload) {
        if (state.mode === "create") {
            create.mutate({ orgId: state.orgId, payload });
        }

        if (state.mode === "edit") {
            update.mutate({ orgId: state.orgId, collectionId: state.collection.id, payload });
        }
    }

    const { collection } = state;

    return (
        <Dialog
            open={open}
            modal
            onOpenChange={handleClose}
        >
            <DialogFixedContent className="sm:max-w-[560px]">
                <DialogHeader>
                    <DialogTitle>{collection ? `Edit ${collection.name}` : "New collection"}</DialogTitle>
                    <DialogDescription>
                        A collection holds hosts the organization shares, sealed with a key only the shellwing apps of
                        those granted hold.
                    </DialogDescription>
                </DialogHeader>

                {open && (
                    <EditCollectionForm
                        key={collection?.id ?? "new"}
                        defaultValues={{
                            name: collection?.name ?? "",
                            description: collection?.description ?? "",
                        }}
                        nameServerError={nameServerError}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        footer={
                            <DialogActionFooter className="flex items-center justify-end gap-3">
                                <Button
                                    type="button"
                                    variant="outline"
                                    className="min-w-[100px]"
                                    disabled={isPending}
                                    onClick={handleClose}
                                >
                                    Cancel
                                </Button>
                                <Button
                                    type="submit"
                                    variant="default"
                                    className="min-w-[100px]"
                                    isLoading={isPending}
                                >
                                    {isEdit ? "Save" : "Create"}
                                </Button>
                            </DialogActionFooter>
                        }
                    />
                )}
            </DialogFixedContent>
        </Dialog>
    );
}
