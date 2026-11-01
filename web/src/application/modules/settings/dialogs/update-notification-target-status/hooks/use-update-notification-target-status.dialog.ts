import type { UpdateNotificationTargetStatusDialogOptions } from "../types";

import { useUpdateNotificationTargetStatusDialogState } from "./use-update-notification-target-status.dialog.state";

function createHook() {
    return function useUpdateNotificationTargetStatusDialog(
        props: UpdateNotificationTargetStatusDialogOptions["props"] = {},
    ) {
        const { state, props: _, ...actions } = useUpdateNotificationTargetStatusDialogState();

        return {
            state,
            actions: {
                open: (id: string, options: UpdateNotificationTargetStatusDialogOptions = {}) => {
                    actions.open(id, {
                        ...options,
                        props: { ...props, ...options.props },
                    });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useUpdateNotificationTargetStatusDialog = createHook();
