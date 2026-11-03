import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useLicenseApi } from "~/system-settings/api/hooks";
import type { License_FindOne_Res } from "~/system-settings/api/services/license-services";
import { QK } from "~/system-settings/data/constants";

type FindOneRes = License_FindOne_Res;

function useFindOne(options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useLicenseApi();

    return useQuery({
        queryKey: [QK["system.license.find-one"]],
        queryFn: ({ signal }) => queries.findOne(signal),
        ...options,
    });
}

export const LicenseQueries = Object.freeze({
    useFindOne,
});
