import { useMemo } from "react";

import { Plus } from "lucide-react";
import { NotificationQueries } from "~/settings/data/queries";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

import { SettingsCreateButton } from "../settings-create-button";

import { NotificationTargetTableDefs } from "./notification-target-table.defs";

export function SettingsNotificationTargetTable() {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { navigate } = useAppNavigate();

    const { data: { data: notificationTargets, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        NotificationQueries.useFindManyPaginated({ pagination, sorting, search });
    const columns = useMemo(() => NotificationTargetTableDefs.columns(), []);

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <div className="flex flex-wrap gap-3">
                        <SettingsCreateButton
                            onClick={() => {
                                navigate.modules(ROUTE.settings.notificationTargets.create.$route);
                            }}
                        >
                            <Plus className="size-4" />
                            New Notification Target
                        </SettingsCreateButton>
                    </div>
                }
            />
            <DataTable
                columns={columns}
                data={notificationTargets}
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
