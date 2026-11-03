import { type UseQueryOptions, useQuery } from "@tanstack/react-query";

import { useOrgsApi } from "@application/shared/api";
import { type Orgs_FindMany_Res } from "@application/shared/api/services";
import { QK } from "@application/shared/data/constants";

type FindManyRes = Orgs_FindMany_Res;

function useFindMany(options: Omit<UseQueryOptions<FindManyRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useOrgsApi();

    return useQuery({
        queryKey: [QK["orgs.find-many"]],
        queryFn: ({ signal }) => queries.findMany(signal),
        ...options,
    });
}

export const OrgsQueries = Object.freeze({
    useFindMany,
});
