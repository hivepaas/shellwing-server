import { memo } from "react";

import type { Collection } from "~/organization/domain";

function View({ collection }: Props) {
    return (
        <div className="flex max-w-[360px] flex-col">
            <span className="font-medium">{collection.name}</span>
            {collection.description ? (
                <span className="text-xs whitespace-pre-line text-muted-foreground">{collection.description}</span>
            ) : null}
        </div>
    );
}

interface Props {
    collection: Collection;
}

export const CollectionNameCell = memo(View);
