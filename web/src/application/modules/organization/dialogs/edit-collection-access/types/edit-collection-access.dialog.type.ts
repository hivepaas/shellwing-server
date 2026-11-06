import { type Collection } from "~/organization/domain";

export interface EditCollectionAccessDialogState {
    state:
        | {
              mode: "open";
              orgId: string;
              collection: Collection;
          }
        | {
              mode: "closed";
              orgId: null;
              collection: null;
          };
}

export interface EditCollectionAccessDialogOptions {
    props?: {
        onSuccess?: () => void;
        onClose?: () => void;
    };
}
