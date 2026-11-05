import { Skeleton } from "@components/ui/skeleton";
import { dashedBorderBox, listBox } from "@lib/styles";
import { cn } from "@lib/utils";

import { useProfileContext } from "@application/shared/context";
import { useAdministeredOrg } from "@application/shared/hooks";
import { PageNoAccess } from "@application/shared/pages";

import { MembersSection } from "../building-blocks";

export function MembersRoute() {
    const { org, isLoading, error } = useAdministeredOrg();
    const myUserId = useProfileContext(state => state.profile?.id ?? "");

    if (isLoading) {
        return (
            <div className={cn(listBox, "flex flex-col gap-4")}>
                <Skeleton className="h-9 w-full" />
                <Skeleton className="h-32 w-full" />
            </div>
        );
    }

    if (error) {
        return (
            <div className={cn(listBox)}>
                <div className={cn(dashedBorderBox)}>
                    <span className="font-semibold text-destructive">Failed to get your organizations:</span>{" "}
                    {error.message}
                </div>
            </div>
        );
    }

    // Only an organization's owners and admins manage its members.
    if (org === null) {
        return <PageNoAccess />;
    }

    return (
        <div className={cn(listBox, "flex flex-col gap-8")}>
            <MembersSection
                orgId={org.id}
                myUserId={myUserId}
                myRole={org.role}
            />
        </div>
    );
}
