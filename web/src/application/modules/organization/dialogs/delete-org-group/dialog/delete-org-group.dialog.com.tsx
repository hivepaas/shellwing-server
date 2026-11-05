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
import { OrgGroupsCommands } from "~/organization/data/commands";

import { useDeleteOrgGroupDialogState } from "../hooks";

export function DeleteOrgGroupDialog() {
    const { state, props: dialogProps, ...actions } = useDeleteOrgGroupDialogState();
    const { mutate: deleteGroup, isPending, reset } = OrgGroupsCommands.useDeleteOne();

    const open = state.mode !== "closed";
    const { group } = state;
    const memberCount = group?.members.length ?? 0;

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

        const { name } = state.group;

        deleteGroup(
            { orgId: state.orgId, groupId: state.group.id },
            {
                onSuccess: () => {
                    toast.success(`Group ${name} deleted`);
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
                    <DialogTitle>Delete {group?.name ?? ""}?</DialogTitle>
                    <DialogDescription>This cannot be undone.</DialogDescription>
                </DialogHeader>

                <DialogBody className="flex flex-col gap-3 text-sm">
                    <p>
                        {memberCount === 0
                            ? "It has no members."
                            : `Its ${memberCount === 1 ? "member loses" : `${memberCount} members lose`} the group's accounts from their next certificate, unless they have them otherwise.`}
                    </p>
                    <p className="text-muted-foreground">Certificates already signed work until they end.</p>
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
