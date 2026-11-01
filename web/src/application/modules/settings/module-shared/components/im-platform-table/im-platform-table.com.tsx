import { useMemo } from "react";

import { Plus } from "lucide-react";
import { ImServiceQueries } from "~/settings/data/queries";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

import { SettingsCreateButton } from "../settings-create-button";

import { ImPlatformTableDefs } from "./im-platform-table.defs";

export function SettingsImPlatformTable() {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { navigate } = useAppNavigate();

    const { data: { data: imPlatformItems, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        ImServiceQueries.useFindManyPaginated({ pagination, sorting, search });
    const columns = useMemo(() => ImPlatformTableDefs.columns(), []);

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <div className="flex flex-wrap gap-3">
                        <SettingsCreateButton
                            onClick={() => {
                                navigate.modules(ROUTE.settings.imPlatforms.create.$route);
                            }}
                        >
                            <Plus className="size-4" />
                            New IM Platform
                        </SettingsCreateButton>
                    </div>
                }
            />
            <DataTable
                columns={columns}
                data={imPlatformItems}
                pageSize={pagination.size}
                manualPagination
                totalCount={meta.page.total}
                manualSorting
                enableSorting
                enablePagination
                isLoading={isFetching}
                onPaginationChange={setPagination}
                onSortingChange={setSorting}
            />
        </div>
    );
}
