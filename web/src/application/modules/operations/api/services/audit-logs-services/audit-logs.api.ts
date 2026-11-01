import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    AuditLogs_FindManyPaginated_Req,
    AuditLogs_FindManyPaginated_Res,
    AuditLogs_FindOneById_Req,
    AuditLogs_FindOneById_Res,
    AuditLogs_FindTypes_Res,
} from "./audit-logs.api.contracts";
import type { AuditLogsApiValidator } from "./audit-logs.api.validator";

const AUDIT_LOGS_ENDPOINT = "/system/audit-logs";

export class AuditLogsApi extends BaseApi {
    public constructor(private readonly validator: AuditLogsApiValidator) {
        super();
    }

    async findManyPaginated(
        request: AuditLogs_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<AuditLogs_FindManyPaginated_Res, Error>> {
        const {
            pagination,
            sorting,
            search,
            type,
            section,
            source,
            result,
            actorID,
            resourceID,
            fromDate,
            toDate,
            scopeOnly,
        } = request.data;
        const query = this.queryBuilder.getInstance();

        query
            .pagination(pagination)
            .sorting(sorting)
            .search(search)
            .filterBy({
                type,
                section,
                source,
                result,
                actorId: actorID,
                resourceId: resourceID,
                fromDate: fromDate ? [fromDate] : undefined,
                toDate: toDate ? [toDate] : undefined,
                scopeOnly: scopeOnly ? [true] : undefined,
            });

        return lastValueFrom(
            from(
                this.client.v1.get(AUDIT_LOGS_ENDPOINT, {
                    params: query.build(),
                    signal,
                }),
            ).pipe(
                map(this.validator.findManyPaginated),
                map(data => Ok(data)),
                catchError(err => of(Err(parseApiError(err)))),
            ),
        );
    }

    async findOneById(
        request: AuditLogs_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<AuditLogs_FindOneById_Res, Error>> {
        const { itemID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`${AUDIT_LOGS_ENDPOINT}/${itemID}`, {
                    signal,
                }),
            ).pipe(
                map(this.validator.findOneById),
                map(data => Ok(data)),
                catchError(err => of(Err(parseApiError(err)))),
            ),
        );
    }

    async findTypes(signal?: AbortSignal): Promise<Result<AuditLogs_FindTypes_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.get(`${AUDIT_LOGS_ENDPOINT}/types`, {
                    signal,
                }),
            ).pipe(
                map(this.validator.findTypes),
                map(data => Ok(data)),
                catchError(err => of(Err(parseApiError(err)))),
            ),
        );
    }
}
