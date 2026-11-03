import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useSshPrincipalsApi } from "~/organization/api/hooks";
import type {
    SshPrincipals_FindOne_Req,
    SshPrincipals_FindOne_Res,
} from "~/organization/api/services/ssh-principals-services";
import { QK } from "~/organization/data/constants";

type FindOneReq = SshPrincipals_FindOne_Req["data"];
type FindOneRes = SshPrincipals_FindOne_Res;

function useFindOne(request: FindOneReq, options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useSshPrincipalsApi();

    return useQuery({
        queryKey: [QK["organization.ssh-principals.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

export const SshPrincipalsQueries = Object.freeze({
    useFindOne,
});
