import { type OrgMember } from "~/organization/domain";

export interface RemoveOrgMemberDialogState {
    state:
        | {
              mode: "open";
              orgId: string;
              member: OrgMember;
              /** The signed-in user themselves: they leave the organization. */
              isSelf: boolean;
          }
        | {
              mode: "closed";
              orgId: null;
              member: null;
              isSelf: false;
          };
}

export interface RemoveOrgMemberDialogOptions {
    props?: {
        onSuccess?: () => void;
        onClose?: () => void;
    };
}
