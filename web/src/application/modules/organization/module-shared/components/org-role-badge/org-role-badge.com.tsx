import { memo } from "react";

import { Badge, type BadgeTone } from "@components/ui/badge";

import { EOrgRole } from "@application/shared/enums";

const ROLE_LABELS: Partial<Record<string, string>> = {
    [EOrgRole.Owner]: "Owner",
    [EOrgRole.Admin]: "Admin",
    [EOrgRole.Member]: "Member",
};

const ROLE_TONES: Partial<Record<string, BadgeTone>> = {
    [EOrgRole.Owner]: "orange",
    [EOrgRole.Admin]: "amber",
    [EOrgRole.Member]: "purple",
};

function View({ role }: Props) {
    return <Badge tone={ROLE_TONES[role] ?? "neutral"}>{ROLE_LABELS[role] ?? role}</Badge>;
}

interface Props {
    role: string;
}

export const OrgRoleBadge = memo(View);
