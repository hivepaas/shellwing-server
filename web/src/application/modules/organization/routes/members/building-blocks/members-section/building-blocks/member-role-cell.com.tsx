import { memo } from "react";

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { toast } from "sonner";
import { OrgMembersCommands } from "~/organization/data/commands";
import { EOrgMemberStatus, type OrgMember } from "~/organization/domain";
import type { OrgMemberAbilities } from "~/organization/module-shared/utils";

import { EOrgRole } from "@application/shared/enums";

const ROLE_OPTIONS: { value: EOrgRole; label: string }[] = [
    { value: EOrgRole.Owner, label: "Owner" },
    { value: EOrgRole.Admin, label: "Admin" },
    { value: EOrgRole.Member, label: "Member" },
];

function isOrgRole(value: string): value is EOrgRole {
    return ROLE_OPTIONS.some(option => option.value === value);
}

function View({ orgId, member, abilities }: Props) {
    const { mutate: updateRole, isPending } = OrgMembersCommands.useUpdateRole({
        onSuccess: (_, request) => {
            const label = ROLE_OPTIONS.find(option => option.value === request.payload.role)?.label ?? "";
            toast.success(`${member.fullName || member.username} is now ${label.toLowerCase()}`);
        },
    });

    // A removed member holds no role until added back.
    if (member.status === EOrgMemberStatus.Removed) {
        return <span className="text-muted-foreground">-</span>;
    }

    return (
        <Select
            value={member.role}
            disabled={!abilities.canChangeRole || isPending}
            onValueChange={value => {
                if (value === member.role || !isOrgRole(value)) {
                    return;
                }

                updateRole({ orgId, userId: member.userId, payload: { role: value } });
            }}
        >
            <SelectTrigger
                size="sm"
                className="w-[120px]"
                aria-label={`Role of ${member.username}`}
            >
                <SelectValue />
            </SelectTrigger>
            <SelectContent>
                {ROLE_OPTIONS.map(option => (
                    <SelectItem
                        key={option.value}
                        value={option.value}
                        disabled={option.value === EOrgRole.Owner && !abilities.canMakeOwner}
                    >
                        {option.label}
                    </SelectItem>
                ))}
            </SelectContent>
        </Select>
    );
}

interface Props {
    orgId: string;
    member: OrgMember;
    abilities: OrgMemberAbilities;
}

export const MemberRoleCell = memo(View);
