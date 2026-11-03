import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useSshCaApi } from "~/organization/api/hooks";
import type { SshCas_FindMany_Req, SshCas_FindMany_Res } from "~/organization/api/services/ssh-ca-services";
import { QK } from "~/organization/data/constants";

type FindManyReq = SshCas_FindMany_Req["data"];
type FindManyRes = SshCas_FindMany_Res;

function useFindMany(request: FindManyReq, options: Omit<UseQueryOptions<FindManyRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useSshCaApi();

    return useQuery({
        queryKey: [QK["organization.ssh-cas.find-many"], request],
        queryFn: ({ signal }) => queries.findMany(request, signal),
        ...options,
    });
}

export const SshCaQueries = Object.freeze({
    useFindMany,
});
