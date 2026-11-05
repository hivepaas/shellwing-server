import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";

/**
 * Account names as chips, or what to say when there are none.
 */
function View({ principals, emptyText = "None", className }: Props) {
    if (principals.length === 0) {
        return <span className="text-sm text-muted-foreground">{emptyText}</span>;
    }

    return (
        <div className={cn("flex flex-wrap gap-1.5", className)}>
            {principals.map(name => (
                <Badge
                    key={name}
                    tone="neutral"
                    className="font-mono"
                >
                    {name}
                </Badge>
            ))}
        </div>
    );
}

interface Props {
    principals: readonly string[];
    emptyText?: string;
    className?: string;
}

export const SshPrincipalBadges = memo(View);
