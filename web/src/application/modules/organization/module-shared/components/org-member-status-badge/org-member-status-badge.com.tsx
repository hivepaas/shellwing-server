import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge, type StatusMeaning, statusClassName } from "@components/ui/badge";
import { EOrgMemberStatus } from "~/organization/domain";

const STATUS_LABELS: Partial<Record<string, string>> = {
    [EOrgMemberStatus.Active]: "Active",
    [EOrgMemberStatus.Invited]: "Invited",
    [EOrgMemberStatus.Disabled]: "Disabled",
};

const STATUS_MEANINGS: Partial<Record<string, StatusMeaning>> = {
    [EOrgMemberStatus.Active]: "success",
    [EOrgMemberStatus.Invited]: "attention",
    [EOrgMemberStatus.Disabled]: "failure",
};

function View({ status }: Props) {
    return (
        <Badge className={cn(statusClassName(STATUS_MEANINGS[status] ?? "info"))}>
            {STATUS_LABELS[status] ?? status}
        </Badge>
    );
}

interface Props {
    status: string;
}

export const OrgMemberStatusBadge = memo(View);
