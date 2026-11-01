import { useMemo } from "react";

import { Plus } from "lucide-react";
import { KeyAuthQueries } from "~/settings/data/queries";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

import { SettingsCreateButton } from "../settings-create-button";

import { KeyAuthTableDefs } from "./key-auth-table.defs";

export function SettingsKeyAuthTable() {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { navigate } = useAppNavigate();

    const { data: { data: keyAuthItems, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        KeyAuthQueries.useFindManyPaginated({ pagination, sorting, search });
    const columns = useMemo(() => KeyAuthTableDefs.columns(), []);

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <div className="flex flex-wrap gap-3">
                        <SettingsCreateButton
                            onClick={() => {
                                navigate.modules(ROUTE.settings.keyAuth.create.$route);
                            }}
                        >
                            <Plus className="size-4" />
                            New Key Auth
                        </SettingsCreateButton>
                    </div>
                }
            />
            <DataTable
                columns={columns}
                data={keyAuthItems}
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
