import { memo } from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    EditSshPrincipalsDialog,
    RemoveOrgMemberDialog,
    ReplaceSshCaDialog,
    useEditSshPrincipalsDialogState,
    useRemoveOrgMemberDialogState,
    useReplaceSshCaDialogState,
} from "~/organization/dialogs";

function View() {
    const location = useLocation();
    const replaceSshCaDialog = useReplaceSshCaDialogState();
    const editSshPrincipalsDialog = useEditSshPrincipalsDialogState();
    const removeOrgMemberDialog = useRemoveOrgMemberDialogState();

    useUpdateEffect(() => {
        replaceSshCaDialog.destroy();
        editSshPrincipalsDialog.destroy();
        removeOrgMemberDialog.destroy();
    }, [location]);

    return (
        <>
            <ReplaceSshCaDialog />
            <EditSshPrincipalsDialog />
            <RemoveOrgMemberDialog />
        </>
    );
}

export const OrganizationDialogsContainer = memo(View);
