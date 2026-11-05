import { type OrgGroup } from "~/organization/domain";

export interface EditOrgGroupMembersDialogState {
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

export interface EditOrgGroupMembersDialogOptions {
    props?: {
        onSuccess?: (group: OrgGroup) => void;
        onClose?: () => void;
    };
}
