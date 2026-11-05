import { memo } from "react";

import { Badge } from "@components/ui/badge";
import type { OrgMember } from "~/organization/domain";

function View({ member, isSelf }: Props) {
    return (
        <div className="flex flex-col">
            <div className="flex items-center gap-2">
                <span className="font-medium">{member.fullName || member.username}</span>
                {isSelf && <Badge tone="neutral">You</Badge>}
            </div>
            {member.fullName ? <span className="text-xs text-muted-foreground">{member.username}</span> : null}
        </div>
    );
}

interface Props {
    member: OrgMember;
    isSelf: boolean;
}

export const MemberPersonCell = memo(View);
