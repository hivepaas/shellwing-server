import { Button } from "@components/ui/button";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { format } from "date-fns";
import { Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { LicenseCommands } from "~/system-settings/data/commands";
import { ELicenseStatus, type License } from "~/system-settings/domain";
import { CopyableValue, LicenseStatusBadge } from "~/system-settings/module-shared/components";

import { ContentBlock, InfoBlock, PopConfirm } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction } from "@application/shared/permissions";

const TITLE_WIDTH = 220;

function formatDate(date: Date) {
    return format(date, "yyyy-MM-dd");
}

function seatsText({ usedUsers, maxUsers }: License) {
    if (maxUsers === null) {
        return `${usedUsers} taken, no limit`;
    }

    return `${usedUsers} of ${maxUsers} taken`;
}

function statusNote(license: License) {
    switch (license.status) {
        case ELicenseStatus.None:
            return `No key is installed: up to ${license.freeUsers} users are free.`;
        case ELicenseStatus.Grace:
            return license.graceEndsAt
                ? `The key has expired. Its limit holds until ${formatDate(license.graceEndsAt)}, then the free limit is back.`
                : "The key has expired. Its limit holds for its grace, then the free limit is back.";
        case ELicenseStatus.Expired:
            return `The key is past its grace: the free limit of ${license.freeUsers} users is back.`;
        case ELicenseStatus.Invalid:
            return `The installed key no longer verifies: the free limit of ${license.freeUsers} users holds.`;
        case ELicenseStatus.OtherInstallation:
            return `The installed key is for another server: the free limit of ${license.freeUsers} users holds.`;
        default:
            return null;
    }
}

export function LicenseSummary({ license }: Props) {
    const { mutate: removeLicense, isPending: isRemoving } = LicenseCommands.useRemoveOne({
        onSuccess: () => {
            toast.success("License key removed. The free limit is back.");
        },
    });

    const hasKey = license.status !== ELicenseStatus.None;
    const isFull = license.maxUsers !== null && license.usedUsers >= license.maxUsers;
    const note = statusNote(license);

    return (
        <ContentBlock label="License">
            <div className="flex flex-col gap-6">
                <p className="text-sm text-muted-foreground">
                    Up to {license.freeUsers} users are free, and a license key lifts the limit. A user holds a seat
                    while active or invited; a disabled or deleted one does not. Past the limit no new user can be made
                    or let back in, and that is all: nobody is removed or locked out, and every feature keeps working,
                    also when a license ends.
                </p>

                <InfoBlock
                    title="Server ID"
                    description="A key is issued for this ID: give it when you get one."
                    titleWidth={TITLE_WIDTH}
                >
                    <CopyableValue
                        value={license.installationId}
                        label="Copy the server ID"
                    />
                </InfoBlock>

                <InfoBlock
                    title="Status"
                    titleWidth={TITLE_WIDTH}
                >
                    <div className="flex flex-col gap-2 text-sm">
                        <div className="flex flex-wrap items-center gap-2">
                            <LicenseStatusBadge status={license.status} />
                            {license.trial && license.status !== ELicenseStatus.Trial ? (
                                <span className="text-muted-foreground">Trial key</span>
                            ) : null}
                        </div>
                        {note && <span className="text-muted-foreground">{note}</span>}
                    </div>
                </InfoBlock>

                <InfoBlock
                    title="Seats"
                    titleWidth={TITLE_WIDTH}
                >
                    <div className="flex flex-col gap-2 text-sm">
                        <span className="font-medium">{seatsText(license)}</span>
                        {isFull && (
                            <div className={cn(dashedBorderBox)}>
                                <span className="font-semibold text-orange-500">Every seat is taken:</span> no new user
                                can be invited or let back in. Those already here keep working.
                            </div>
                        )}
                    </div>
                </InfoBlock>

                {hasKey && (
                    <InfoBlock
                        title="Installed key"
                        titleWidth={TITLE_WIDTH}
                    >
                        <div className="flex flex-col gap-3 text-sm">
                            <dl className="grid grid-cols-[max-content_1fr] gap-x-4 gap-y-1">
                                {license.customerName && (
                                    <>
                                        <dt className="text-muted-foreground">Customer</dt>
                                        <dd>{license.customerName}</dd>
                                    </>
                                )}
                                {license.plan && (
                                    <>
                                        <dt className="text-muted-foreground">Plan</dt>
                                        <dd>{license.plan}</dd>
                                    </>
                                )}
                                <dt className="text-muted-foreground">Expires</dt>
                                <dd>{license.expiresAt ? formatDate(license.expiresAt) : "Never"}</dd>
                                {license.graceEndsAt && (
                                    <>
                                        <dt className="text-muted-foreground">Grace ends</dt>
                                        <dd>{formatDate(license.graceEndsAt)}</dd>
                                    </>
                                )}
                            </dl>
                            <PermissionTooltipAction
                                id={MODULE_IDS.System}
                                action="delete"
                            >
                                {({ isDenied }) => (
                                    <PopConfirm
                                        title="Remove the license key?"
                                        description={`The free limit of ${license.freeUsers} users is back. Nobody is removed or locked out: past the limit, only new users cannot be added.`}
                                        confirmText="Remove"
                                        variant="destructive"
                                        align="start"
                                        onConfirm={() => {
                                            removeLicense();
                                        }}
                                    >
                                        <Button
                                            type="button"
                                            variant="destructiveOutline"
                                            className="w-fit"
                                            isLoading={isRemoving}
                                            disabled={isDenied}
                                        >
                                            {!isRemoving && <Trash2Icon className="size-4" />}
                                            Remove key
                                        </Button>
                                    </PopConfirm>
                                )}
                            </PermissionTooltipAction>
                        </div>
                    </InfoBlock>
                )}
            </div>
        </ContentBlock>
    );
}

interface Props {
    license: License;
}
