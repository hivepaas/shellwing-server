import { useMemo } from "react";

import { Plus } from "lucide-react";
import { EmailQueries } from "~/settings/data/queries";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

import { SettingsCreateButton } from "../settings-create-button";

import { EmailAccountTableDefs } from "./email-account-table.defs";

export function SettingsEmailAccountsTable() {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { navigate } = useAppNavigate();

    const { data: { data: emailAccountItems, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        EmailQueries.useFindManyPaginated({ pagination, sorting, search });
    const columns = useMemo(() => EmailAccountTableDefs.columns(), []);

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <div className="flex flex-wrap gap-3">
                        <SettingsCreateButton
                            onClick={() => {
                                navigate.modules(ROUTE.settings.emailAccounts.create.$route);
                            }}
                        >
                            <Plus className="size-4" />
                            New Email Account
                        </SettingsCreateButton>
                    </div>
                }
            />
            <DataTable
                columns={columns}
                data={emailAccountItems}
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
