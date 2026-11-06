import { Button } from "@components/ui/button";
import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogDescription,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { toast } from "sonner";
import { CollectionsCommands } from "~/organization/data/commands";

import { useDeleteCollectionDialogState } from "../hooks";

export function DeleteCollectionDialog() {
    const { state, props: dialogProps, ...actions } = useDeleteCollectionDialogState();
    const { mutate: deleteCollection, isPending, reset } = CollectionsCommands.useDeleteOne();

    const open = state.mode !== "closed";
    const { collection } = state;

    function handleClose() {
        if (isPending) {
            return;
        }

        reset();
        actions.close();
        dialogProps?.onClose?.();
    }

    function handleConfirm() {
        if (state.mode !== "open") {
            return;
        }

        const { name } = state.collection;

        deleteCollection(
            { orgId: state.orgId, collectionId: state.collection.id },
            {
                onSuccess: () => {
                    toast.success(`Collection ${name} deleted`);
                    actions.close();
                    dialogProps?.onSuccess?.();
                },
            },
        );
    }

    return (
        <Dialog
            open={open}
            modal
            onOpenChange={handleClose}
        >
            <DialogFixedContent className="sm:max-w-[480px]">
                <DialogHeader>
                    <DialogTitle>Delete {collection?.name ?? ""}?</DialogTitle>
                    <DialogDescription>This cannot be undone.</DialogDescription>
                </DialogHeader>

                <DialogBody className="flex flex-col gap-3 text-sm">
                    <p>
                        Nobody syncs it from now on, and the hosts it holds on this server are deleted, with
                        everyone&apos;s access to it.
                    </p>
                </DialogBody>

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
                        type="button"
                        variant="destructive"
                        className="min-w-[100px]"
                        isLoading={isPending}
                        onClick={handleConfirm}
                    >
                        Delete
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}
