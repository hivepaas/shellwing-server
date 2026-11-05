import { useState } from "react";

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
import { SshPrincipalsCommands } from "~/organization/data/commands";
import { SshPrincipalsQueries } from "~/organization/data/queries";
import { formatSshPrincipals } from "~/organization/module-shared/utils";

import { isSshPrincipalInvalidException } from "@infrastructure/api";

import { EditSshPrincipalsForm } from "../form";
import { useEditSshPrincipalsDialogState } from "../hooks";

export function EditSshPrincipalsDialog() {
    const { state, props: dialogProps, ...actions } = useEditSshPrincipalsDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const open = state.mode !== "closed";
    const orgId = state.orgId ?? "";
    const userId = state.member?.userId ?? "";

    // What they have now, asked for afresh: the list it was opened from may be a while old.
    const { data, isError } = SshPrincipalsQueries.useFindOne({ orgId, userId }, { enabled: open });

    const {
        mutate: updatePrincipals,
        isPending,
        error,
        reset,
    } = SshPrincipalsCommands.useUpdateOne({
        onSuccess: response => {
            toast.success("Accounts saved. They apply from the member's next certificate.");
            setHasChanges(false);
            actions.close();
            dialogProps?.onSuccess?.(response.data);
        },
    });

    const serverError = error !== null && isSshPrincipalInvalidException(error) ? error.message : null;

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

    function onSubmit(principals: string[]) {
        if (state.mode !== "open") {
            return;
        }

        updatePrincipals({
            orgId: state.orgId,
            userId: state.member.userId,
            payload: { principals },
        });
    }

    const { member } = state;
    const name = member ? member.fullName || member.username : "";
    const groupNames = member ? member.groups.map(group => group.name) : [];

    return (
        <Dialog
            open={open}
            modal
            onOpenChange={handleClose}
        >
            <DialogFixedContent className="sm:max-w-[560px]">
                <DialogHeader>
                    <DialogTitle>Accounts of {name}</DialogTitle>
                    <DialogDescription>
                        The accounts {member ? <span className="font-mono">{member.username}</span> : null} may log in
                        as on the servers that trust this organization&apos;s certificate authority. Only their own are
                        edited here: those from their groups are set under Groups.
                    </DialogDescription>
                    {groupNames.length > 0 && (
                        <p className="text-sm text-muted-foreground">Their groups: {groupNames.join(", ")}</p>
                    )}
                </DialogHeader>

                {data === undefined && isError && (
                    <DialogBody className="text-sm text-muted-foreground">
                        The accounts of this member could not be read. Close this and try again.
                    </DialogBody>
                )}

                {data === undefined && !isError && (
                    <DialogBody className="flex flex-col gap-3">
                        <Skeleton className="h-4 w-24" />
                        <Skeleton className="h-20 w-full" />
                    </DialogBody>
                )}

                {data !== undefined && (
                    <EditSshPrincipalsForm
                        key={userId}
                        defaultValues={{ principals: formatSshPrincipals(data.data) }}
                        serverError={serverError}
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
