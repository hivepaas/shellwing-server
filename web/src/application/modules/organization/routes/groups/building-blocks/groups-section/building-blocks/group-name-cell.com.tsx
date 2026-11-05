import { memo } from "react";

import type { OrgGroup } from "~/organization/domain";

function View({ group }: Props) {
    return (
        <div className="flex max-w-[360px] flex-col">
            <span className="font-medium">{group.name}</span>
            {group.description ? (
                <span className="text-xs whitespace-pre-line text-muted-foreground">{group.description}</span>
            ) : null}
        </div>
    );
}

interface Props {
    group: OrgGroup;
}

export const GroupNameCell = memo(View);
