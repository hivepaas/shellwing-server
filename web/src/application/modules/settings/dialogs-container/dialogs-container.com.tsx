import { memo } from "react";

import { useLocation, useUpdateEffect } from "react-use";
import {
    UpdateCloudStorageStatusDialog,
    UpdateEmailAccountStatusDialog,
    UpdateImPlatformStatusDialog,
    UpdateKeyAuthStatusDialog,
    UpdateNotificationTargetStatusDialog,
    UpdateOAuthStatusDialog,
    useUpdateCloudStorageStatusDialogState,
    useUpdateEmailAccountStatusDialogState,
    useUpdateImPlatformStatusDialogState,
    useUpdateKeyAuthStatusDialogState,
    useUpdateNotificationTargetStatusDialogState,
    useUpdateOAuthStatusDialogState,
} from "~/settings/dialogs";

function View() {
    const location = useLocation();
    const updateImPlatformStatusDialog = useUpdateImPlatformStatusDialogState();
    const updateEmailAccountStatusDialog = useUpdateEmailAccountStatusDialogState();
    const updateCloudStorageStatusDialog = useUpdateCloudStorageStatusDialogState();
    const updateOAuthStatusDialog = useUpdateOAuthStatusDialogState();
    const updateNotificationTargetStatusDialog = useUpdateNotificationTargetStatusDialogState();
    const updateKeyAuthStatusDialog = useUpdateKeyAuthStatusDialogState();

    useUpdateEffect(() => {
        updateImPlatformStatusDialog.destroy();
        updateEmailAccountStatusDialog.destroy();
        updateCloudStorageStatusDialog.destroy();
        updateOAuthStatusDialog.destroy();
        updateNotificationTargetStatusDialog.destroy();
        updateKeyAuthStatusDialog.destroy();
    }, [location]);

    return (
        <>
            <UpdateImPlatformStatusDialog />
            <UpdateEmailAccountStatusDialog />
            <UpdateCloudStorageStatusDialog />
            <UpdateOAuthStatusDialog />
            <UpdateNotificationTargetStatusDialog />
            <UpdateKeyAuthStatusDialog />
        </>
    );
}

export const SettingsDialogsContainer = memo(View);
