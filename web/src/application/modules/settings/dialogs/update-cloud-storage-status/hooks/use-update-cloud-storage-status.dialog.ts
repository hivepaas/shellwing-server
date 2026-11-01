import type { UpdateCloudStorageStatusDialogOptions } from "../types";

import { useUpdateCloudStorageStatusDialogState } from "./use-update-cloud-storage-status.dialog.state";

function createHook() {
    return function useUpdateCloudStorageStatusDialog(props: UpdateCloudStorageStatusDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useUpdateCloudStorageStatusDialogState();

        return {
            state,
            actions: {
                open: (id: string, options: UpdateCloudStorageStatusDialogOptions = {}) => {
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

export const useUpdateCloudStorageStatusDialog = createHook();
