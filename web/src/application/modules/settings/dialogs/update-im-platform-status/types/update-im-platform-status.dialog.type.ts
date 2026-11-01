export interface UpdateImPlatformStatusDialogState {
    state: { mode: "open"; id: string } | { mode: "closed" };
}

export interface UpdateImPlatformStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
