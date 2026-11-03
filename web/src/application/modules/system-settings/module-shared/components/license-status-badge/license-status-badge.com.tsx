import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge, type StatusMeaning, statusClassName } from "@components/ui/badge";
import { ELicenseStatus } from "~/system-settings/domain";

const STATUS_LABELS: Partial<Record<string, string>> = {
    [ELicenseStatus.None]: "No key",
    [ELicenseStatus.Valid]: "Valid",
    [ELicenseStatus.Trial]: "Trial",
    [ELicenseStatus.Grace]: "In grace",
    [ELicenseStatus.Expired]: "Expired",
    [ELicenseStatus.Invalid]: "Invalid",
    [ELicenseStatus.OtherInstallation]: "For another server",
};

const STATUS_MEANINGS: Partial<Record<string, StatusMeaning>> = {
    [ELicenseStatus.None]: "info",
    [ELicenseStatus.Valid]: "success",
    [ELicenseStatus.Trial]: "success",
    [ELicenseStatus.Grace]: "attention",
    [ELicenseStatus.Expired]: "failure",
    [ELicenseStatus.Invalid]: "failure",
    [ELicenseStatus.OtherInstallation]: "failure",
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

export const LicenseStatusBadge = memo(View);
