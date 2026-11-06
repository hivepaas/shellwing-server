import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useCollectionsApi } from "~/organization/api/hooks";
import type {
    Collections_FindMany_Req,
    Collections_FindMany_Res,
    Collections_FindOne_Req,
    Collections_FindOne_Res,
} from "~/organization/api/services/collections-services";
import { QK } from "~/organization/data/constants";

type FindManyReq = Collections_FindMany_Req["data"];
type FindManyRes = Collections_FindMany_Res;

function useFindMany(request: FindManyReq, options: Omit<UseQueryOptions<FindManyRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useCollectionsApi();

    return useQuery({
        queryKey: [QK["organization.collections.find-many"], request],
        queryFn: ({ signal }) => queries.findMany(request, signal),
        ...options,
    });
}

type FindOneReq = Collections_FindOne_Req["data"];
type FindOneRes = Collections_FindOne_Res;

function useFindOne(request: FindOneReq, options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useCollectionsApi();

    return useQuery({
        queryKey: [QK["organization.collections.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

export const CollectionsQueries = Object.freeze({
    useFindMany,
    useFindOne,
});
