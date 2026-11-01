export interface UpdateEmailAccountStatusDialogState {
    state: { mode: "open"; id: string } | { mode: "closed" };
}

export interface UpdateEmailAccountStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
