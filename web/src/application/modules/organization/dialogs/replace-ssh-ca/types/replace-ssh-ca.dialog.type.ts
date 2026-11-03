export interface ReplaceSshCaDialogState {
    state:
        | {
              mode: "open";
              orgId: string;
          }
        | {
              mode: "closed";
              orgId: null;
          };
}

export interface ReplaceSshCaDialogOptions {
    props?: {
        onSuccess?: () => void;
        onClose?: () => void;
    };
}
