import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useOrgGroupsApi } from "~/organization/api/hooks";
import type { OrgGroups_FindMany_Req, OrgGroups_FindMany_Res } from "~/organization/api/services/org-groups-services";
import { QK } from "~/organization/data/constants";

type FindManyReq = OrgGroups_FindMany_Req["data"];
type FindManyRes = OrgGroups_FindMany_Res;

function useFindMany(request: FindManyReq, options: Omit<UseQueryOptions<FindManyRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useOrgGroupsApi();

    return useQuery({
        queryKey: [QK["organization.groups.find-many"], request],
        queryFn: ({ signal }) => queries.findMany(request, signal),
        ...options,
    });
}

export const OrgGroupsQueries = Object.freeze({
    useFindMany,
});
