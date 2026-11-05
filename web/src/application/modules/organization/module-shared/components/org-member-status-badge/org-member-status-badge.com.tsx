import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge, type StatusMeaning, statusClassName } from "@components/ui/badge";
import { EOrgMemberStatus } from "~/organization/domain";

import { EUserStatus } from "@application/shared/enums";

const STATUS_LABELS: Partial<Record<string, string>> = {
    [EOrgMemberStatus.Active]: "Active",
    [EOrgMemberStatus.Suspended]: "Suspended",
    [EOrgMemberStatus.Removed]: "Removed",
};

const STATUS_MEANINGS: Partial<Record<string, StatusMeaning>> = {
    [EOrgMemberStatus.Active]: "success",
    [EOrgMemberStatus.Suspended]: "attention",
    [EOrgMemberStatus.Removed]: "inactive",
};

/** The user account's own status, shown beside the member's when it keeps them out: the member's alone does not say. */
const ACCOUNT_STATUS_LABELS: Partial<Record<string, string>> = {
    [EUserStatus.Pending]: "Invite pending",
    [EUserStatus.Disabled]: "Account disabled",
};

const ACCOUNT_STATUS_MEANINGS: Partial<Record<string, StatusMeaning>> = {
    [EUserStatus.Pending]: "waiting",
    [EUserStatus.Disabled]: "failure",
};

function View({ status, accountStatus, className }: Props) {
    const showAccountStatus = accountStatus !== undefined && accountStatus !== EUserStatus.Active;

    return (
        <div className={cn("flex flex-wrap items-center gap-1.5", className)}>
            <Badge className={cn(statusClassName(STATUS_MEANINGS[status] ?? "info"))}>
                {STATUS_LABELS[status] ?? status}
            </Badge>
            {showAccountStatus && (
                <Badge className={cn(statusClassName(ACCOUNT_STATUS_MEANINGS[accountStatus] ?? "info"))}>
                    {ACCOUNT_STATUS_LABELS[accountStatus] ?? `Account ${accountStatus}`}
                </Badge>
            )}
        </div>
    );
}

interface Props {
    /** The member's status in the organization. */
    status: string;
    /** Their user account's status: shown too when it is not active. */
    accountStatus?: string;
    className?: string;
}

export const OrgMemberStatusBadge = memo(View);
