import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge, statusClassName } from "@components/ui/badge";
import type { Collection } from "~/organization/domain";

/**
 * The state of a collection's key: its version, how many people wait to be given it, and whether it has to be replaced.
 */
function View({ collection, className }: Props) {
    const { keyVersion, waiting, needsNewKey } = collection;

    return (
        <div className={cn("flex flex-wrap items-center gap-1.5", className)}>
            {keyVersion === 0 ? (
                <span className="text-sm text-muted-foreground">No key yet</span>
            ) : (
                <span className="text-sm">Key v{keyVersion}</span>
            )}
            {waiting > 0 && (
                <Badge className={cn(statusClassName("waiting"))}>
                    {waiting === 1 ? "1 waiting for the key" : `${waiting} waiting for the key`}
                </Badge>
            )}
            {needsNewKey && <Badge className={cn(statusClassName("attention"))}>Needs a new key</Badge>}
        </div>
    );
}

interface Props {
    collection: Pick<Collection, "keyVersion" | "waiting" | "needsNewKey">;
    className?: string;
}

export const CollectionKeyBadges = memo(View);
