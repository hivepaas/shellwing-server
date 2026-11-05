import { useMemo, useState } from "react";

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
import { Skeleton } from "@components/ui/skeleton";
import { toast } from "sonner";
import { OrgGroupsCommands } from "~/organization/data/commands";
import { OrgMembersQueries } from "~/organization/data/queries";
import { EOrgMemberStatus } from "~/organization/domain";

import { EditOrgGroupMembersForm } from "../form";
import { useEditOrgGroupMembersDialogState } from "../hooks";

export function EditOrgGroupMembersDialog() {
    const { state, props: dialogProps, ...actions } = useEditOrgGroupMembersDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const open = state.mode !== "closed";
    const orgId = state.orgId ?? "";

    const { data, isError } = OrgMembersQueries.useFindMany({ orgId }, { enabled: open });

    // A removed member cannot be in a group until added back.
    const members = useMemo(
        () => (data?.data ?? []).filter(member => member.status !== EOrgMemberStatus.Removed),
        [data],
    );

    const {
        mutate: setMembers,
        isPending,
        reset,
    } = OrgGroupsCommands.useSetMembers({
        onSuccess: response => {
            toast.success("Members saved. The group's accounts apply from their next certificate.");
            setHasChanges(false);
            actions.close();
            dialogProps?.onSuccess?.(response.data);
        },
    });

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
        reset();
        actions.close();
        dialogProps?.onClose?.();
    }

    function onSubmit(userIds: string[]) {
        if (state.mode !== "open") {
            return;
        }

        setMembers({
            orgId: state.orgId,
            groupId: state.group.id,
            payload: { userIds },
        });
    }

    const { group } = state;
    const defaultUserIds = useMemo(() => (group?.members ?? []).map(member => member.userId), [group]);

    return (
        <Dialog
            open={open}
            modal
            onOpenChange={handleClose}
        >
            <DialogFixedContent className="sm:max-w-[520px]">
                <DialogHeader>
                    <DialogTitle>Members of {group?.name ?? ""}</DialogTitle>
                    <DialogDescription>Each of them may log in as the group&apos;s accounts.</DialogDescription>
                </DialogHeader>

                {data === undefined && isError && (
                    <DialogBody className="text-sm text-muted-foreground">
                        The members of the organization could not be read. Close this and try again.
                    </DialogBody>
                )}

                {data === undefined && !isError && (
                    <DialogBody className="flex flex-col gap-3">
                        <Skeleton className="h-9 w-full" />
                        <Skeleton className="h-40 w-full" />
                    </DialogBody>
                )}

                {data !== undefined && group !== null && (
                    <EditOrgGroupMembersForm
                        key={group.id}
                        members={members}
                        defaultUserIds={defaultUserIds}
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
                                    Save
                                </Button>
                            </DialogActionFooter>
                        }
                    />
                )}
            </DialogFixedContent>
        </Dialog>
    );
}
