export interface UpdateNotificationTargetStatusDialogState {
    state: { mode: "open"; id: string } | { mode: "closed" };
}

export interface UpdateNotificationTargetStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
