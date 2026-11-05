import { type OrgGroup } from "~/organization/domain";

export interface EditOrgGroupDialogState {
    state:
        | {
              mode: "create";
              orgId: string;
              group: null;
          }
        | {
              mode: "edit";
              orgId: string;
              group: OrgGroup;
          }
        | {
              mode: "closed";
              orgId: null;
              group: null;
          };
}

export interface EditOrgGroupDialogOptions {
    props?: {
        onSuccess?: (group: OrgGroup) => void;
        onClose?: () => void;
    };
}
