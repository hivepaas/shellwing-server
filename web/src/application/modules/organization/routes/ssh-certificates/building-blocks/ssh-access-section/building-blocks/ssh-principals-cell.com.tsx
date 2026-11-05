import { memo } from "react";

import { Badge } from "@components/ui/badge";
import { Tooltip, TooltipContent, TooltipTrigger } from "@components/ui/tooltip";
import type { GroupSshPrincipal } from "~/organization/module-shared/utils";

/**
 * A member's own accounts, then those they get from their groups - muted, with the groups on hover.
 */
function View({ principals, groupPrincipals }: Props) {
    if (principals.length === 0 && groupPrincipals.length === 0) {
        return <span className="text-muted-foreground">None</span>;
    }

    return (
        <div className="flex flex-wrap gap-1.5">
            {principals.map(name => (
                <Badge
                    key={name}
                    tone="neutral"
                    className="font-mono"
                >
                    {name}
                </Badge>
            ))}
            {groupPrincipals.map(({ name, groupNames }) => (
                <Tooltip key={name}>
                    <TooltipTrigger asChild>
                        <Badge
                            variant="outline"
                            tabIndex={0}
                            className="cursor-default border-dashed font-mono text-muted-foreground"
                            aria-label={`${name}, from ${groupNames.join(", ")}`}
                        >
                            {name}
                        </Badge>
                    </TooltipTrigger>
                    <TooltipContent>From {groupNames.join(", ")}</TooltipContent>
                </Tooltip>
            ))}
        </div>
    );
}

interface Props {
    /** Their own. */
    principals: readonly string[];
    /** From their groups, but not their own. */
    groupPrincipals: readonly GroupSshPrincipal[];
}

export const SshPrincipalsCell = memo(View);
