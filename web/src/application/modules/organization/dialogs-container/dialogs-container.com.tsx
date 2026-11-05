import { memo } from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    DeleteOrgGroupDialog,
    EditOrgGroupDialog,
    EditOrgGroupMembersDialog,
    EditSshPrincipalsDialog,
    RemoveOrgMemberDialog,
    ReplaceSshCaDialog,
    useDeleteOrgGroupDialogState,
    useEditOrgGroupDialogState,
    useEditOrgGroupMembersDialogState,
    useEditSshPrincipalsDialogState,
    useRemoveOrgMemberDialogState,
    useReplaceSshCaDialogState,
} from "~/organization/dialogs";

function View() {
    const location = useLocation();
    const replaceSshCaDialog = useReplaceSshCaDialogState();
    const editSshPrincipalsDialog = useEditSshPrincipalsDialogState();
    const removeOrgMemberDialog = useRemoveOrgMemberDialogState();
    const editOrgGroupDialog = useEditOrgGroupDialogState();
    const editOrgGroupMembersDialog = useEditOrgGroupMembersDialogState();
    const deleteOrgGroupDialog = useDeleteOrgGroupDialogState();

    useUpdateEffect(() => {
        replaceSshCaDialog.destroy();
        editSshPrincipalsDialog.destroy();
        removeOrgMemberDialog.destroy();
        editOrgGroupDialog.destroy();
        editOrgGroupMembersDialog.destroy();
        deleteOrgGroupDialog.destroy();
    }, [location]);

    return (
        <>
            <ReplaceSshCaDialog />
            <EditSshPrincipalsDialog />
            <RemoveOrgMemberDialog />
            <EditOrgGroupDialog />
            <EditOrgGroupMembersDialog />
            <DeleteOrgGroupDialog />
        </>
    );
}

export const OrganizationDialogsContainer = memo(View);
