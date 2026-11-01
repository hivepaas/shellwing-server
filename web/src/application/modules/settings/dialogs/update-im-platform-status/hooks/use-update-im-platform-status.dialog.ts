import type { UpdateImPlatformStatusDialogOptions } from "../types";

import { useUpdateImPlatformStatusDialogState } from "./use-update-im-platform-status.dialog.state";

function createHook() {
    return function useUpdateImPlatformStatusDialog(props: UpdateImPlatformStatusDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useUpdateImPlatformStatusDialogState();

        return {
            state,
            actions: {
                open: (id: string, options: UpdateImPlatformStatusDialogOptions = {}) => {
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

export const useUpdateImPlatformStatusDialog = createHook();
