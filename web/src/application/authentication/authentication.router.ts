import {
    CreateAccountRoute,
    ForgotPasswordRoute,
    ResetPasswordRoute,
    SignInRoute,
    SignUpRoute,
    SsoRoute,
    SsoSignUpRoute,
    TwoFaRoute,
} from "@/application/authentication/routes";
import { ROUTE } from "@/application/shared/constants";
import { type RouteObject } from "react-router";

export const authenticationRouter: RouteObject = {
    children: [
        {
            path: ROUTE.auth.signIn.$pattern,
            Component: SignInRoute,
        },
        {
            path: ROUTE.auth.twoFA.$pattern,
            Component: TwoFaRoute,
        },
        {
            path: ROUTE.auth.resetPassword.$pattern,
            Component: ResetPasswordRoute,
        },
        {
            path: ROUTE.auth.forgotPassword.$pattern,
            Component: ForgotPasswordRoute,
        },
        {
            path: ROUTE.auth.signUp.$pattern,
            Component: SignUpRoute,
        },
        {
            path: ROUTE.auth.createAccount.$pattern,
            Component: CreateAccountRoute,
        },
        {
            path: ROUTE.auth.sso.success.$pattern,
            Component: SsoRoute,
        },
        {
            path: ROUTE.auth.sso.signUp.$pattern,
            Component: SsoSignUpRoute,
        },
    ],
} as const;
