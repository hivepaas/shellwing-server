import { memo } from "react";

import { Tooltip, TooltipContent, TooltipTrigger } from "@components/ui/tooltip";
import type { OrgGroupMember } from "~/organization/domain";

/** How many names are shown before the rest go into "+N more". */
const SHOWN_NAMES = 3;

function getName(member: OrgGroupMember) {
    return member.fullName || member.username;
}

function View({ members }: Props) {
    if (members.length === 0) {
        return <span className="text-muted-foreground">None</span>;
    }

    const shown = members.slice(0, SHOWN_NAMES);
    const rest = members.length - shown.length;

    return (
        <div className="flex max-w-[320px] flex-wrap items-center gap-x-1 text-sm">
            <span>{shown.map(getName).join(", ")}</span>
            {rest > 0 && (
                <Tooltip>
                    <TooltipTrigger asChild>
                        <button
                            type="button"
                            className="cursor-default text-muted-foreground underline decoration-dotted underline-offset-2"
                        >
                            +{rest} more
                        </button>
                    </TooltipTrigger>
                    <TooltipContent className="max-w-[280px]">
                        {members.length} members: {members.map(getName).join(", ")}
                    </TooltipContent>
                </Tooltip>
            )}
        </div>
    );
}

interface Props {
    members: readonly OrgGroupMember[];
}

export const GroupMembersCell = memo(View);
