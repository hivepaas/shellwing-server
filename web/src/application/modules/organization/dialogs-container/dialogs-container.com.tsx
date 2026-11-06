import { memo } from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    DeleteCollectionDialog,
    DeleteOrgGroupDialog,
    EditCollectionAccessDialog,
    EditCollectionDialog,
    EditOrgGroupDialog,
    EditOrgGroupMembersDialog,
    EditSshPrincipalsDialog,
    RemoveOrgMemberDialog,
    ReplaceSshCaDialog,
    useDeleteCollectionDialogState,
    useDeleteOrgGroupDialogState,
    useEditCollectionAccessDialogState,
    useEditCollectionDialogState,
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
    const editCollectionDialog = useEditCollectionDialogState();
    const editCollectionAccessDialog = useEditCollectionAccessDialogState();
    const deleteCollectionDialog = useDeleteCollectionDialogState();

    useUpdateEffect(() => {
        replaceSshCaDialog.destroy();
        editSshPrincipalsDialog.destroy();
        removeOrgMemberDialog.destroy();
        editOrgGroupDialog.destroy();
        editOrgGroupMembersDialog.destroy();
        deleteOrgGroupDialog.destroy();
        editCollectionDialog.destroy();
        editCollectionAccessDialog.destroy();
        deleteCollectionDialog.destroy();
    }, [location]);

    return (
        <>
            <ReplaceSshCaDialog />
            <EditSshPrincipalsDialog />
            <RemoveOrgMemberDialog />
            <EditOrgGroupDialog />
            <EditOrgGroupMembersDialog />
            <DeleteOrgGroupDialog />
            <EditCollectionDialog />
            <EditCollectionAccessDialog />
            <DeleteCollectionDialog />
        </>
    );
}

export const OrganizationDialogsContainer = memo(View);
