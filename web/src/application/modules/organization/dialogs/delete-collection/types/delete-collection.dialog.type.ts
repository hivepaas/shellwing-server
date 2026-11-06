import { type Collection } from "~/organization/domain";

export interface DeleteCollectionDialogState {
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

export interface DeleteCollectionDialogOptions {
    props?: {
        onSuccess?: () => void;
        onClose?: () => void;
    };
}
