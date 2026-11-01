export interface UpdateKeyAuthStatusDialogState {
    state: { mode: "open"; id: string } | { mode: "closed" };
}

export interface UpdateKeyAuthStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
