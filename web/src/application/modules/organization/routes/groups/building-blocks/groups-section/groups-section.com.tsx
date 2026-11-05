import { useMemo } from "react";

import { Button } from "@components/ui/button";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { Plus } from "lucide-react";
import { OrgGroupsQueries } from "~/organization/data/queries";
import { useEditOrgGroupDialog } from "~/organization/dialogs";

import { ContentBlock, TableActions } from "@application/shared/components";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

import { GroupsTableDefs } from "./groups-table.defs";

export function GroupsSection({ orgId }: Props) {
    const { search, setSearch } = useTableState();
    const editDialog = useEditOrgGroupDialog();
    const { data, isFetching, error } = OrgGroupsQueries.useFindMany({ orgId });
    const columns = useMemo(() => GroupsTableDefs.columns(orgId), [orgId]);

    const all = useMemo(() => data?.data ?? [], [data]);
    const hasFailed = data === undefined && !isFetching;
    const isEmpty = data !== undefined && all.length === 0;

    // The server sends every group at once, sorted by name: the search narrows them here.
    const groups = useMemo(() => {
        const term = search.trim().toLowerCase();

        if (!term) {
            return all;
        }

        return all.filter(
            group =>
                group.name.toLowerCase().includes(term) ||
                group.description.toLowerCase().includes(term) ||
                group.sshPrincipals.some(name => name.toLowerCase().includes(term)) ||
                group.members.some(
                    member =>
                        member.username.toLowerCase().includes(term) || member.fullName.toLowerCase().includes(term),
                ),
        );
    }, [all, search]);

    function openCreate() {
        editDialog.actions.create(orgId);
    }

    const newGroupButton = (
        <Button
            type="button"
            onClick={openCreate}
        >
            <Plus className="size-4" />
            New group
        </Button>
    );

    return (
        <ContentBlock label="Groups">
            <div className="flex flex-col gap-4">
                <p className="text-sm text-muted-foreground">
                    A group grants SSH accounts to several members at once. Its accounts are each member&apos;s, on top
                    of their own. A change applies from each member&apos;s next certificate.
                </p>

                {hasFailed && (
                    <div className={cn(dashedBorderBox)}>
                        <span className="font-semibold text-destructive">Failed to get the groups:</span>{" "}
                        {error?.message ?? "Unknown error"}
                    </div>
                )}

                {isEmpty && (
                    <div className={cn(dashedBorderBox, "flex flex-col items-start gap-3")}>
                        <span>No groups yet. Make one to grant the same accounts to several members.</span>
                        {newGroupButton}
                    </div>
                )}

                {!hasFailed && !isEmpty && (
                    <>
                        <TableActions
                            search={{ value: search, onChange: setSearch }}
                            renderActions={newGroupButton}
                        />
                        <DataTable
                            columns={columns}
                            data={groups}
                            getRowId={group => group.id}
                            enablePagination
                            enableSorting
                            isLoading={isFetching}
                        />
                    </>
                )}
            </div>
        </ContentBlock>
    );
}

interface Props {
    orgId: string;
}
