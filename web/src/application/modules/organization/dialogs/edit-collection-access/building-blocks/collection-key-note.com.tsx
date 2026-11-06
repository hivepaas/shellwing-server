import { memo } from "react";

import { cn } from "@lib/utils";
import { Info, TriangleAlert } from "lucide-react";
import type { Collection } from "~/organization/domain";
import { CollectionKeyBadges } from "~/organization/module-shared/components";

/**
 * The state of the collection's key, and how people come to hold it: never through the server or the web.
 */
function View({ collection, className }: Props) {
    return (
        <div className={cn("flex flex-col gap-3", className)}>
            <div className="flex flex-wrap items-center gap-2">
                <span className="text-sm font-medium">Key</span>
                <CollectionKeyBadges collection={collection} />
            </div>

            {collection.needsNewKey && (
                <div className="flex items-start gap-2 rounded-md border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-sm text-amber-800 dark:text-amber-300">
                    <TriangleAlert className="mt-0.5 size-4 shrink-0" />
                    <span>
                        Somebody who held the key has lost access. The shellwing app of someone with Manage replaces the
                        key the next time it opens the collection.
                    </span>
                </div>
            )}

            <div className="flex items-start gap-2 rounded-md border bg-muted px-3 py-2 text-sm text-muted-foreground">
                <Info className="mt-0.5 size-4 shrink-0" />
                <div className="flex flex-col gap-1.5">
                    <span>
                        The server and the web never hold a collection&apos;s key. People granted access wait until the
                        shellwing app of someone who holds the key (an owner, or someone with Manage) compares their
                        fingerprint and gives it to them.
                    </span>
                    <span>
                        A collection with no key yet gets one the first time the app of someone with Manage opens it.
                    </span>
                </div>
            </div>
        </div>
    );
}

interface Props {
    collection: Collection;
    className?: string;
}

export const CollectionKeyNote = memo(View);
