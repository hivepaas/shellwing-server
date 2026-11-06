import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useUserIdentityApi } from "~/user-management/api";
import { type UserIdentity_FindMine_Res } from "~/user-management/api/services";
import { QK } from "~/user-management/data/constants";

/**
 * Find the signed-in person's identity query
 */
type FindMineRes = UserIdentity_FindMine_Res;

type FindMineOptions = Omit<UseQueryOptions<FindMineRes>, "queryKey" | "queryFn">;

function useFindMine(options: FindMineOptions = {}) {
    const { queries } = useUserIdentityApi();

    return useQuery({
        queryKey: [QK["users.identity.find-mine"]],
        queryFn: ({ signal }) => queries.findMine(signal),
        ...options,
    });
}

export const UserIdentityQueries = Object.freeze({
    useFindMine,
});
