import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useOrgMembersApi } from "~/organization/api/hooks";
import type {
    OrgMembers_FindMany_Req,
    OrgMembers_FindMany_Res,
} from "~/organization/api/services/org-members-services";
import { QK } from "~/organization/data/constants";

type FindManyReq = OrgMembers_FindMany_Req["data"];
type FindManyRes = OrgMembers_FindMany_Res;

function useFindMany(request: FindManyReq, options: Omit<UseQueryOptions<FindManyRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useOrgMembersApi();

    return useQuery({
        queryKey: [QK["organization.members.find-many"], request],
        queryFn: ({ signal }) => queries.findMany(request, signal),
        ...options,
    });
}

export const OrgMembersQueries = Object.freeze({
    useFindMany,
});
