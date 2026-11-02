import { z } from "zod";

import { ESecuritySettings, EUserRole } from "@application/shared/enums";

import {
    type Auth_AuthorizeDevice_Res,
    type Auth_CompleteSsoSignup_Res,
    type Auth_ForgotPassword_Res,
    type Auth_GetLoginOptions_Res,
    type Auth_GetSignupOptions_Res,
    type Auth_GetSsoSignupInfo_Res,
    type Auth_RequestSignUp_Res,
    type Auth_ResetPassword_Res,
    type Auth_Send2FAToken_Res,
    type Auth_SignIn2FA_Res,
    type Auth_SignIn_Res,
    type Auth_SignUp_Res,
    type Auth_ValidateInviteToken_Res,
    type Auth_ValidateResetToken_Res,
} from "@application/authentication/api/services";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * Sign in API response schema
 */
const SignInSchema = z.object({
    data: z.union([
        z.object({
            session: z.object({
                accessToken: z.string(),
            }),
            nextStep: z.string().optional(),
        }),

        z.object({
            mfaType: z.string(),
            mfaToken: z.string(),
        }),
    ]),
});

/**
 * Send 2FA token API response schema
 */
const Send2FATokenSchema = z.object({
    data: z.object({
        requestBlockingDuration: z.number(),
    }),
});

/**
 * Sign in 2FA API response schema
 */
const SignIn2FASchema = z.object({
    data: z.object({
        session: z.object({
            accessToken: z.string(),
        }),
    }),
});

/**
 * Forgot password API response schema
 */
const AuthorizeDeviceSchema = z.object({
    data: z.object({
        redirectUrl: z.string(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const ForgotPasswordSchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

/**
 * Get login options API response schema
 */
const GetLoginOptionsSchema = z.object({
    data: z.array(
        z.object({
            authURL: z.string(),
            name: z.string(),
            type: z.string(),
            icon: z.string(),
        }),
    ),
});

/**
 * Get sign-up options API response schema
 */
const GetSignupOptionsSchema = z.object({
    data: z.object({
        selfSignupEnabled: z.boolean(),
    }),
});

/**
 * The account a sign-in with a provider offers API response schema
 */
const GetSsoSignupInfoSchema = z.object({
    data: z.object({
        email: z.string(),
        fullName: z
            .string()
            .nullish()
            .transform(fullName => fullName?.trim() ?? ""),
    }),
});

/**
 * The account made from a sign-in with a provider API response schema: its
 * session, as a password's sign-in answers it
 */
const CompleteSsoSignupSchema = z.object({
    data: z.object({
        session: z.object({
            accessToken: z.string(),
        }),
    }),
});

/**
 * The name the account was invited or signed up with. An older server does
 * not send it, and a blank one is no name.
 */
const CandidateFullNameSchema = z
    .string()
    .nullish()
    .transform(fullName => fullName?.trim() ?? "");

/**
 * Validate invite token API response schema
 */
const ValidateInviteTokenSchema = z.object({
    data: z.discriminatedUnion("securityOption", [
        z.object({
            email: z.string().trim().email(),
            fullName: CandidateFullNameSchema,
            role: z.nativeEnum(EUserRole),
            accessExpireAt: z.coerce.date(),
            securityOption: z.literal(ESecuritySettings.PasswordOnly),
        }),
        z.object({
            email: z.string().trim().email(),
            fullName: CandidateFullNameSchema,
            role: z.nativeEnum(EUserRole),
            accessExpireAt: z.coerce.date(),
            mfaTotpSecret: z.string(),
            qrCode: z.object({
                dataBase64: z.string(),
            }),
            securityOption: z.literal(ESecuritySettings.Password2FA),
        }),
        z.object({
            email: z.string().trim().email(),
            fullName: CandidateFullNameSchema,
            role: z.nativeEnum(EUserRole),
            accessExpireAt: z.coerce.date(),
            securityOption: z.literal(ESecuritySettings.EnforceSSO),
        }),
    ]),
});

export class AuthApiValidator {
    /**
     * Validate and transform the sign-up API response
     */
    signUp = (_: ApiHttpResponse): Auth_SignUp_Res => {
        return {
            data: {
                type: "success",
            },
        };
    };

    /**
     * Validate and transform the sign-up request API response. The answer is
     * the same whether or not the address has an account, so there is
     * nothing in it to read.
     */
    requestSignUp = (_: ApiHttpResponse): Auth_RequestSignUp_Res => {
        return {
            data: {
                type: "success",
            },
        };
    };

    /**
     * Validate and transform the sign in API response
     */
    signIn = (response: ApiHttpResponse): Auth_SignIn_Res => {
        const { data } = parseApiResponse({
            response,
            schema: SignInSchema,
        });

        if ("session" in data) {
            if ("nextStep" in data && data.nextStep === "NextMfaSetup") {
                return {
                    data: {
                        type: "mfa-setup-required",
                        token: data.session.accessToken,
                    },
                };
            }

            return {
                data: {
                    type: "success",
                    token: data.session.accessToken,
                },
            };
        }

        return {
            data: {
                type: "mfa-required",
                mfaToken: data.mfaToken,
            },
        };
    };

    /**
     * Validate and transform the send 2FA token API response
     */
    send2FAToken = (response: ApiHttpResponse): Auth_Send2FAToken_Res => {
        const { data } = parseApiResponse({
            response,
            schema: Send2FATokenSchema,
        });

        return {
            data,
        };
    };

    /**
     * Validate and transform the sign in 2FA API response
     */
    signIn2FA = (response: ApiHttpResponse): Auth_SignIn2FA_Res => {
        const { data } = parseApiResponse({
            response,
            schema: SignIn2FASchema,
        });

        return {
            data: {
                token: data.session.accessToken,
            },
        };
    };

    /**
     * Validate and transform the authorize device API response
     */
    authorizeDevice = (response: ApiHttpResponse): Auth_AuthorizeDevice_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: AuthorizeDeviceSchema,
        });

        return { data, meta };
    };

    /**
     * Validate and transform the forgot password API response
     */
    forgotPassword = (response: ApiHttpResponse): Auth_ForgotPassword_Res => {
        const { meta } = parseApiResponse({
            response,
            schema: ForgotPasswordSchema,
        });

        return {
            data: {
                type: "success",
            },
            meta,
        };
    };

    /**
     * Validate and transform the reset token validation API response
     */
    validateResetToken = (_: ApiHttpResponse): Auth_ValidateResetToken_Res => {
        return {
            data: {
                type: "success",
            },
        };
    };

    /**
     * Validate and transform the reset password API response
     */
    resetPassword = (_: ApiHttpResponse): Auth_ResetPassword_Res => {
        return {
            data: {
                type: "success",
            },
        };
    };

    /**
     * Validate and transform the get login options API response
     */
    getLoginOptions = (response: ApiHttpResponse): Auth_GetLoginOptions_Res => {
        const { data } = parseApiResponse({
            response,
            schema: GetLoginOptionsSchema,
        });

        return {
            data,
        };
    };

    /**
     * Validate and transform the get sign-up options API response
     */
    getSignupOptions = (response: ApiHttpResponse): Auth_GetSignupOptions_Res => {
        const { data } = parseApiResponse({
            response,
            schema: GetSignupOptionsSchema,
        });

        return {
            data,
        };
    };

    /**
     * Validate and transform the account a sign-in with a provider offers
     */
    getSsoSignupInfo = (response: ApiHttpResponse): Auth_GetSsoSignupInfo_Res => {
        const { data } = parseApiResponse({
            response,
            schema: GetSsoSignupInfoSchema,
        });

        return {
            data,
        };
    };

    /**
     * Validate and transform the account made from a sign-in with a provider
     */
    completeSsoSignup = (response: ApiHttpResponse): Auth_CompleteSsoSignup_Res => {
        const { data } = parseApiResponse({
            response,
            schema: CompleteSsoSignupSchema,
        });

        return {
            data: {
                type: "success",
                token: data.session.accessToken,
            },
        };
    };

    /**
     * Validate and transform the validate invite token API response
     */
    validateInviteToken = (response: ApiHttpResponse): Auth_ValidateInviteToken_Res => {
        const { data } = parseApiResponse({
            response,
            schema: ValidateInviteTokenSchema,
        });

        if (
            data.securityOption === ESecuritySettings.PasswordOnly ||
            data.securityOption === ESecuritySettings.EnforceSSO
        ) {
            return {
                data: {
                    candidate: {
                        email: data.email,
                        fullName: data.fullName,
                        role: data.role,
                        securityOption: data.securityOption,
                        accessExpiration: data.accessExpireAt,
                    },
                },
            };
        }
        return {
            data: {
                candidate: {
                    email: data.email,
                    fullName: data.fullName,
                    role: data.role,
                    securityOption: data.securityOption,
                    accessExpiration: data.accessExpireAt,
                    mfaTotpSecret: data.mfaTotpSecret,
                    qrCode: data.qrCode.dataBase64,
                },
            },
        };
    };
}
