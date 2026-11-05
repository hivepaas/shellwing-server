import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";
import type { OrgGroupRef } from "~/organization/domain";

/**
 * The groups a member is in, by name.
 */
function View({ groups, className }: Props) {
    if (groups.length === 0) {
        return <span className="text-muted-foreground">-</span>;
    }

    return (
        <div className={cn("flex flex-wrap gap-1.5", className)}>
            {groups.map(group => (
                <Badge
                    key={group.id}
                    tone="indigo"
                >
                    {group.name}
                </Badge>
            ))}
        </div>
    );
}

interface Props {
    groups: readonly OrgGroupRef[];
    className?: string;
}

export const OrgGroupBadges = memo(View);
