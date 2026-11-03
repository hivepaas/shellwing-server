import { memo } from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    EditSshPrincipalsDialog,
    ReplaceSshCaDialog,
    useEditSshPrincipalsDialogState,
    useReplaceSshCaDialogState,
} from "~/organization/dialogs";

function View() {
    const location = useLocation();
    const replaceSshCaDialog = useReplaceSshCaDialogState();
    const editSshPrincipalsDialog = useEditSshPrincipalsDialogState();

    useUpdateEffect(() => {
        replaceSshCaDialog.destroy();
        editSshPrincipalsDialog.destroy();
    }, [location]);

    return (
        <>
            <ReplaceSshCaDialog />
            <EditSshPrincipalsDialog />
        </>
    );
}

export const OrganizationDialogsContainer = memo(View);
