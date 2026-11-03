import { memo } from "react";

import { Badge } from "@components/ui/badge";

function View({ principals }: Props) {
    if (principals.length === 0) {
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
        </div>
    );
}

interface Props {
    principals: string[];
}

export const SshPrincipalsCell = memo(View);
