import type { PaginationState, SortingState } from "@infrastructure/data";
import type { AuditLog, AuditLogResult, AuditLogSource } from "~/operations/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type AuditLogs_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    /** Matched against the actor's and the resource's names, the client IP and the remote address. */
    search?: string;
    type?: string[];
    /** Narrows `type`: a section belongs to a type. */
    section?: string[];
    source?: AuditLogSource[];
    result?: AuditLogResult[];
    actorID?: string[];
    resourceID?: string[];
    /** yyyy-MM-dd, inclusive. */
    fromDate?: string;
    /** yyyy-MM-dd, inclusive. */
    toDate?: string;
    /** Only the global entries, leaving out the ones scoped to a user. */
    scopeOnly?: boolean;
}>;

export type AuditLogs_FindManyPaginated_Res = ApiResponsePaginated<AuditLog>;

export type AuditLogs_FindOneById_Req = ApiRequestBase<{
    itemID: string;
}>;

export type AuditLogs_FindOneById_Res = ApiResponseBase<AuditLog>;

export type AuditLogs_FindTypes_Res = ApiResponseBase<string[]>;
