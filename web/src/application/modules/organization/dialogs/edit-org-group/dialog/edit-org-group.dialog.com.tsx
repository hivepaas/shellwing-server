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
import type { OrgGroups_Payload } from "~/organization/api/services/org-groups-services";
import { OrgGroupsCommands } from "~/organization/data/commands";
import type { OrgGroup } from "~/organization/domain";
import { formatSshPrincipals } from "~/organization/module-shared/utils";

import { isAlreadyExistException, isSshPrincipalInvalidException } from "@infrastructure/api";

import { EditOrgGroupForm } from "../form";
import { useEditOrgGroupDialogState } from "../hooks";

export function EditOrgGroupDialog() {
    const { state, props: dialogProps, ...actions } = useEditOrgGroupDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const open = state.mode !== "closed";
    const isEdit = state.mode === "edit";

    function onSuccess(group: OrgGroup) {
        setHasChanges(false);
        actions.close();
        dialogProps?.onSuccess?.(group);
    }

    const create = OrgGroupsCommands.useCreateOne({
        onSuccess: response => {
            toast.success(`Group ${response.data.name} created. Add its members next.`);
            onSuccess(response.data);
        },
    });

    const update = OrgGroupsCommands.useUpdateOne({
        onSuccess: response => {
            toast.success("Group saved. Its members get the change from their next certificate.");
            onSuccess(response.data);
        },
    });

    const { isPending, error } = isEdit ? update : create;

    const nameServerError =
        error !== null && isAlreadyExistException(error) ? "Another group has this name already" : null;
    const principalsServerError = error !== null && isSshPrincipalInvalidException(error) ? error.message : null;

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

    function onSubmit(payload: OrgGroups_Payload) {
        if (state.mode === "create") {
            create.mutate({ orgId: state.orgId, payload });
        }

        if (state.mode === "edit") {
            update.mutate({ orgId: state.orgId, groupId: state.group.id, payload });
        }
    }

    const { group } = state;

    return (
        <Dialog
            open={open}
            modal
            onOpenChange={handleClose}
        >
            <DialogFixedContent className="sm:max-w-[560px]">
                <DialogHeader>
                    <DialogTitle>{group ? `Edit ${group.name}` : "New group"}</DialogTitle>
                    <DialogDescription>Its SSH accounts are each member&apos;s, on top of their own.</DialogDescription>
                </DialogHeader>

                {open && (
                    <EditOrgGroupForm
                        key={group?.id ?? "new"}
                        defaultValues={{
                            name: group?.name ?? "",
                            description: group?.description ?? "",
                            principals: formatSshPrincipals(group?.sshPrincipals ?? []),
                        }}
                        nameServerError={nameServerError}
                        principalsServerError={principalsServerError}
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
