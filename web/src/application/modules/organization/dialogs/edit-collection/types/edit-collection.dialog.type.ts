import { type Collection } from "~/organization/domain";

export interface EditCollectionDialogState {
    state:
        | {
              mode: "create";
              orgId: string;
              collection: null;
          }
        | {
              mode: "edit";
              orgId: string;
              collection: Collection;
          }
        | {
              mode: "closed";
              orgId: null;
              collection: null;
          };
}

export interface EditCollectionDialogOptions {
    props?: {
        onSuccess?: (collection: Collection) => void;
        onClose?: () => void;
    };
}
