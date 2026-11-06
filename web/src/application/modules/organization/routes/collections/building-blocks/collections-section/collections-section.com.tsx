import { useMemo } from "react";

import { Button } from "@components/ui/button";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { Plus } from "lucide-react";
import { CollectionsQueries } from "~/organization/data/queries";
import { useEditCollectionDialog } from "~/organization/dialogs";

import { ContentBlock, TableActions } from "@application/shared/components";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

import { CollectionsTableDefs } from "./collections-table.defs";

export function CollectionsSection({ orgId }: Props) {
    const { search, setSearch } = useTableState();
    const editDialog = useEditCollectionDialog();
    const { data, isFetching, error } = CollectionsQueries.useFindMany({ orgId });
    const columns = useMemo(() => CollectionsTableDefs.columns(orgId), [orgId]);

    const all = useMemo(() => data?.data ?? [], [data]);
    const hasFailed = data === undefined && !isFetching;
    const isEmpty = data !== undefined && all.length === 0;

    // The server sends every collection at once: the search narrows them here.
    const collections = useMemo(() => {
        const term = search.trim().toLowerCase();

        if (!term) {
            return all;
        }

        return all.filter(
            collection =>
                collection.name.toLowerCase().includes(term) || collection.description.toLowerCase().includes(term),
        );
    }, [all, search]);

    function openCreate() {
        editDialog.actions.create(orgId);
    }

    const newCollectionButton = (
        <Button
            type="button"
            onClick={openCreate}
        >
            <Plus className="size-4" />
            New collection
        </Button>
    );

    return (
        <ContentBlock label="Collections">
            <div className="flex flex-col gap-4">
                <p className="text-sm text-muted-foreground">
                    A collection holds hosts the organization shares in the shellwing app. Its hosts are sealed with a
                    key the server and the web never hold: the app of someone who holds it gives it to each person
                    granted access, once they have compared fingerprints.
                </p>

                {hasFailed && (
                    <div className={cn(dashedBorderBox)}>
                        <span className="font-semibold text-destructive">Failed to get the collections:</span>{" "}
                        {error?.message ?? "Unknown error"}
                    </div>
                )}

                {isEmpty && (
                    <div className={cn(dashedBorderBox, "flex flex-col items-start gap-3")}>
                        <span>No collections yet. Make one to share hosts with the people you grant access.</span>
                        {newCollectionButton}
                    </div>
                )}

                {!hasFailed && !isEmpty && (
                    <>
                        <TableActions
                            search={{ value: search, onChange: setSearch }}
                            renderActions={newCollectionButton}
                        />
                        <DataTable
                            columns={columns}
                            data={collections}
                            getRowId={collection => collection.id}
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
