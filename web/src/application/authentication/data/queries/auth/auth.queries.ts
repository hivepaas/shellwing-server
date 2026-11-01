import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useAuthApi } from "@application/authentication/api";
import {
    type Auth_GetLoginOptions_Res,
    type Auth_GetSignupOptions_Res,
    type Auth_GetSsoSignupInfo_Res,
} from "@application/authentication/api/services";
import { QK } from "@application/authentication/data/constants";

/**
 * Get login options
 */
type GetLoginOptionsRes = Auth_GetLoginOptions_Res;

type GetLoginOptionsOptions = Omit<UseQueryOptions<GetLoginOptionsRes>, "queryKey" | "queryFn">;

function useGetLoginOptions(options: GetLoginOptionsOptions = {}) {
    const { queries } = useAuthApi();

    return useQuery({
        queryKey: [QK["auth.get-login-options"]],
        queryFn: ({ signal }) => queries.getLoginOptions(signal),
        placeholderData: keepPreviousData,

        ...options,
    });
}

/**
 * Get sign-up options
 */
type GetSignupOptionsRes = Auth_GetSignupOptions_Res;

type GetSignupOptionsOptions = Omit<UseQueryOptions<GetSignupOptionsRes>, "queryKey" | "queryFn">;

function useGetSignupOptions(options: GetSignupOptionsOptions = {}) {
    const { queries } = useAuthApi();

    return useQuery({
        queryKey: [QK["auth.get-signup-options"]],
        queryFn: ({ signal }) => queries.getSignupOptions(signal),
        // A server setting: it does not change while the page is open.
        staleTime: Infinity,

        ...options,
    });
}

/**
 * The account a sign-in with a provider offers, by the token it handed over
 */
type GetSsoSignupInfoRes = Auth_GetSsoSignupInfo_Res;

type GetSsoSignupInfoOptions = Omit<UseQueryOptions<GetSsoSignupInfoRes>, "queryKey" | "queryFn">;

function useGetSsoSignupInfo(token: string, options: GetSsoSignupInfoOptions = {}) {
    const { queries } = useAuthApi();

    return useQuery({
        queryKey: [QK["auth.get-sso-signup-info"], token],
        queryFn: ({ signal }) => queries.getSsoSignupInfo(token, signal),
        enabled: token !== "",
        // A token that is not valid stays so.
        retry: false,

        ...options,
    });
}

export const AuthQueries = Object.freeze({
    useGetLoginOptions,
    useGetSignupOptions,
    useGetSsoSignupInfo,
});
