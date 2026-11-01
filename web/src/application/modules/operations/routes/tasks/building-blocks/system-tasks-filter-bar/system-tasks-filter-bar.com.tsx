import { cn } from "@/lib/utils";
import { format } from "date-fns";
import { SystemTaskStatus, SystemTaskType } from "~/operations/domain";

import { Badge, statusClassName } from "@/components/ui/badge";
import { DateTimePicker } from "@/components/ui/date-time-picker";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export interface SystemTaskFilterValues {
    scopeOnly?: boolean;
    type?: SystemTaskType;
    status?: SystemTaskStatus;
    fromDate?: string;
    toDate?: string;
}

export interface SystemTasksFilterBarProps {
    filters: SystemTaskFilterValues;
    onChange: (filters: SystemTaskFilterValues) => void;
    className?: string;
}

const ALL_TASK_TYPES = Object.values(SystemTaskType);

const ALL_TASK_STATUSES = [
    SystemTaskStatus.NotStarted,
    SystemTaskStatus.InProgress,
    SystemTaskStatus.Done,
    SystemTaskStatus.Failed,
    SystemTaskStatus.Canceled,
];

const STATUS_LABELS: Record<SystemTaskStatus, string> = {
    [SystemTaskStatus.NotStarted]: "Not Started",
    [SystemTaskStatus.InProgress]: "In-Progress",
    [SystemTaskStatus.Done]: "Done",
    [SystemTaskStatus.Failed]: "Failed",
    [SystemTaskStatus.Canceled]: "Canceled",
};

const STATUS_BADGE_CLASS_NAMES: Record<SystemTaskStatus, string> = {
    [SystemTaskStatus.Done]: statusClassName("success"),
    [SystemTaskStatus.Failed]: statusClassName("failure"),
    [SystemTaskStatus.InProgress]: statusClassName("progress"),
    [SystemTaskStatus.NotStarted]: statusClassName("waiting"),
    [SystemTaskStatus.Canceled]: statusClassName("inactive"),
};

export function SystemTasksFilterBar({ filters, onChange, className }: SystemTasksFilterBarProps) {
    function handleScopeChange(val: string) {
        onChange({
            ...filters,
            scopeOnly: val === "global-only" ? true : undefined,
        });
    }

    function handleTypeChange(val: string) {
        onChange({
            ...filters,
            type: val === "all" ? undefined : (val as SystemTaskType),
        });
    }

    function handleStatusChange(val: string) {
        onChange({
            ...filters,
            status: val === "all" ? undefined : (val as SystemTaskStatus),
        });
    }

    function handleFromDateChange(date: Date | undefined) {
        onChange({
            ...filters,
            fromDate: date ? format(date, "yyyy-MM-dd") : undefined,
        });
    }

    function handleToDateChange(date: Date | undefined) {
        onChange({
            ...filters,
            toDate: date ? format(date, "yyyy-MM-dd") : undefined,
        });
    }

    const fromDateValue = filters.fromDate ? new Date(`${filters.fromDate}T00:00:00`) : undefined;
    const toDateValue = filters.toDate ? new Date(`${filters.toDate}T00:00:00`) : undefined;

    return (
        <div
            className={cn(
                "rounded-lg border border-border/80 bg-card/60 p-3.5 sm:p-4 shadow-2xs backdrop-blur-xs flex flex-col gap-3.5 transition-all duration-200",
                className,
            )}
        >
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3">
                {/* 1. Scope Dropdown */}
                <div className="flex flex-col gap-1.5 min-w-0">
                    <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                        Scope
                    </span>
                    <Select
                        value={filters.scopeOnly ? "global-only" : "all"}
                        onValueChange={handleScopeChange}
                    >
                        <SelectTrigger className="h-9 w-full">
                            <SelectValue placeholder="All Scopes" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="all">All Scopes</SelectItem>
                            <SelectItem value="global-only">Global Only</SelectItem>
                        </SelectContent>
                    </Select>
                </div>

                {/* 2. Type Dropdown */}
                <div className="flex flex-col gap-1.5 min-w-0">
                    <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                        Type
                    </span>
                    <Select
                        value={filters.type ?? "all"}
                        onValueChange={handleTypeChange}
                    >
                        <SelectTrigger className="h-9 w-full">
                            <SelectValue placeholder="All Types" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="all">
                                <span>All Types</span>
                            </SelectItem>
                            {ALL_TASK_TYPES.map(typeItem => (
                                <SelectItem
                                    key={typeItem}
                                    value={typeItem}
                                >
                                    <Badge
                                        variant="outline"
                                        className="font-mono text-xs px-2 py-0.5 rounded-md border-border/70 bg-muted/50 text-foreground"
                                    >
                                        {typeItem}
                                    </Badge>
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                </div>

                {/* 3. Status Dropdown */}
                <div className="flex flex-col gap-1.5 min-w-0">
                    <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                        Status
                    </span>
                    <Select
                        value={filters.status ?? "all"}
                        onValueChange={handleStatusChange}
                    >
                        <SelectTrigger className="h-9 w-full">
                            <SelectValue placeholder="All Statuses" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="all">
                                <span>All Statuses</span>
                            </SelectItem>
                            {ALL_TASK_STATUSES.map(statusItem => (
                                <SelectItem
                                    key={statusItem}
                                    value={statusItem}
                                >
                                    <Badge
                                        className={cn(
                                            "h-6 px-2.5 text-xs font-semibold shadow-2xs",
                                            STATUS_BADGE_CLASS_NAMES[statusItem],
                                        )}
                                    >
                                        {STATUS_LABELS[statusItem]}
                                    </Badge>
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                </div>

                {/* 4. From Date */}
                <div className="flex flex-col gap-1.5 min-w-0">
                    <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                        From Date
                    </span>
                    <DateTimePicker
                        value={fromDateValue}
                        onChange={handleFromDateChange}
                        placeholder="From date"
                        granularity="day"
                        showClearButton
                        className="h-9 text-xs sm:text-sm"
                    />
                </div>

                {/* 5. To Date */}
                <div className="flex flex-col gap-1.5 min-w-0">
                    <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                        To Date
                    </span>
                    <DateTimePicker
                        value={toDateValue}
                        onChange={handleToDateChange}
                        placeholder="To date"
                        granularity="day"
                        showClearButton
                        className="h-9 text-xs sm:text-sm"
                    />
                </div>
            </div>
        </div>
    );
}
