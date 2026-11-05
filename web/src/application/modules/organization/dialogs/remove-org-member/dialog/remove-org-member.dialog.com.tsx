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
import { OrgMembersCommands } from "~/organization/data/commands";
import { EOrgMemberStatus } from "~/organization/domain";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import { useRemoveOrgMemberDialogState } from "../hooks";

export function RemoveOrgMemberDialog() {
    const { state, props: dialogProps, ...actions } = useRemoveOrgMemberDialogState();
    const { navigate } = useAppNavigate();
    const { mutate: updateStatus, isPending, reset } = OrgMembersCommands.useUpdateStatus();

    const open = state.mode !== "closed";
    const { member, isSelf } = state;
    const name = member ? member.fullName || member.username : "";

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

        const leaving = state.isSelf;

        updateStatus(
            { orgId: state.orgId, userId: state.member.userId, payload: { status: EOrgMemberStatus.Removed } },
            {
                onSuccess: () => {
                    toast.success(leaving ? "You left the organization" : `${name} was removed`);
                    actions.close();
                    dialogProps?.onSuccess?.();

                    // Its admin pages are not theirs any more.
                    if (leaving) {
                        void navigate.basic(ROUTE.home.$route, { replace: true, ignorePrevPath: true });
                    }
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
                    <DialogTitle>{isSelf ? "Leave the organization?" : `Remove ${name}?`}</DialogTitle>
                    <DialogDescription>
                        {isSelf ? "You stop being a member." : "They stop being a member."}
                    </DialogDescription>
                </DialogHeader>

                <DialogBody className="flex flex-col gap-3 text-sm">
                    {isSelf ? (
                        <p>
                            You lose your role and your SSH accounts here. An owner or admin can add you back later, as
                            a member.
                        </p>
                    ) : (
                        <p>
                            They lose their role and their SSH accounts here. You can add them back later, as a member.
                        </p>
                    )}
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
                        {isSelf ? "Leave" : "Remove"}
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}
