import { Skeleton } from "@components/ui/skeleton";
import { dashedBorderBox, listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { LicenseQueries } from "~/system-settings/data/queries";
import { ELicenseStatus } from "~/system-settings/domain";

import { InstallLicenseForm, LicenseSummary } from "../building-blocks";

export function SystemLicenseRoute() {
    const { data, isLoading, error } = LicenseQueries.useFindOne();

    if (isLoading) {
        return (
            <div className={cn(listBox, "flex flex-col gap-4")}>
                <Skeleton className="h-9 w-full" />
                <Skeleton className="h-32 w-full" />
            </div>
        );
    }

    if (data === undefined) {
        return (
            <div className={cn(listBox)}>
                <div className={cn(dashedBorderBox)}>
                    <span className="font-semibold text-destructive">Failed to get the license:</span>{" "}
                    {error?.message ?? "Unknown error"}
                </div>
            </div>
        );
    }

    const license = data.data;

    return (
        <div className={cn(listBox, "flex flex-col gap-8")}>
            <LicenseSummary license={license} />
            <InstallLicenseForm hasKey={license.status !== ELicenseStatus.None} />
        </div>
    );
}
