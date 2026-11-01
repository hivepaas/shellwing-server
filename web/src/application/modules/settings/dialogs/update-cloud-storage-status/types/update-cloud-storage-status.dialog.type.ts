export interface UpdateCloudStorageStatusDialogState {
    state: { mode: "open"; id: string } | { mode: "closed" };
}

export interface UpdateCloudStorageStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
