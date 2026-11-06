import { memo } from "react";

import { Badge, type BadgeTone } from "@components/ui/badge";
import { ECollectionPermission } from "~/organization/domain";
import { getCollectionPermissionLabel } from "~/organization/module-shared/constants";

const PERMISSION_TONES: Partial<Record<string, BadgeTone>> = {
    [ECollectionPermission.Connect]: "sky",
    [ECollectionPermission.Edit]: "indigo",
    [ECollectionPermission.Manage]: "violet",
};

/**
 * A permission on a collection: Connect, Edit or Manage.
 */
function View({ permission }: Props) {
    if (!permission) {
        return <span className="text-muted-foreground">-</span>;
    }

    return <Badge tone={PERMISSION_TONES[permission] ?? "neutral"}>{getCollectionPermissionLabel(permission)}</Badge>;
}

interface Props {
    permission: string;
}

export const CollectionPermissionBadge = memo(View);
