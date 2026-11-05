import { type OrgGroup } from "~/organization/domain";

export interface DeleteOrgGroupDialogState {
    state:
        | {
              mode: "open";
              orgId: string;
              group: OrgGroup;
          }
        | {
              mode: "closed";
              orgId: null;
              group: null;
          };
}

export interface DeleteOrgGroupDialogOptions {
    props?: {
        onSuccess?: () => void;
        onClose?: () => void;
    };
}
