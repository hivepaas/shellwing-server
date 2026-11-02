import type { DeviceInfo } from "@infrastructure/device";
import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import {
    type AuthApiValidator,
    type Auth_AuthorizeDevice_Req,
    type Auth_AuthorizeDevice_Res,
    type Auth_CompleteSsoSignup_Req,
    type Auth_CompleteSsoSignup_Res,
    type Auth_ForgotPassword_Req,
    type Auth_ForgotPassword_Res,
    type Auth_GetLoginOptions_Res,
    type Auth_GetSignupOptions_Res,
    type Auth_GetSsoSignupInfo_Req,
    type Auth_GetSsoSignupInfo_Res,
    type Auth_RequestSignUp_Req,
    type Auth_RequestSignUp_Res,
    type Auth_ResetPassword_Req,
    type Auth_ResetPassword_Res,
    type Auth_Send2FAToken_Req,
    type Auth_Send2FAToken_Res,
    type Auth_SignIn2FA_Req,
    type Auth_SignIn2FA_Res,
    type Auth_SignIn_Req,
    type Auth_SignIn_Res,
    type Auth_SignUp_Req,
    type Auth_SignUp_Res,
    type Auth_ValidateInviteToken_Req,
    type Auth_ValidateInviteToken_Res,
    type Auth_ValidateResetToken_Req,
    type Auth_ValidateResetToken_Res,
} from "@application/authentication/api/services";

import { BaseApi, parseApiError } from "@infrastructure/api";

/**
 * The longest `next` the server mails on. A longer one would fail the whole
 * request, and the account matters more than where it lands, so it is left out.
 */
const SIGN_UP_NEXT_MAX_LENGTH = 500;

export class AuthApi extends BaseApi {
    public constructor(
        private readonly validator: AuthApiValidator,
        private readonly device: DeviceInfo,
    ) {
        super();
    }

    /**
     * Sign up
     */
    async signUp(request: Auth_SignUp_Req): Promise<Result<Auth_SignUp_Res, Error>> {
        // const json = this.mapper.signUp.toEmailApi(request.data);
        const { inviteToken, data } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post("/users/signup-complete", {
                    inviteToken,
                    username: data.username,
                    password: data.password,
                    fullName: data.fullName,
                    position: data.position,
                    photo: data.photo ?? {
                        fileName: "",
                        dataBase64: "",
                    },
                    mfaTotpSecret: data.mfaTotpSecret == "" ? undefined : data.mfaTotpSecret,
                    passcode: data.passcode == "" ? undefined : data.passcode,
                }),
            ).pipe(
                map(this.validator.signUp),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Request sign-up: the server emails a link to finish creating the account,
     * and answers the same whether or not the address already has one
     */
    async requestSignUp(request: Auth_RequestSignUp_Req): Promise<Result<Auth_RequestSignUp_Res, Error>> {
        const { email, fullName, next } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post("/users/signup-request", {
                    email,
                    fullName,
                    next: next !== undefined && next.length <= SIGN_UP_NEXT_MAX_LENGTH ? next : undefined,
                }),
            ).pipe(
                map(this.validator.requestSignUp),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Sign in with email and password
     */
    async signIn(request: Auth_SignIn_Req): Promise<Result<Auth_SignIn_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.post(
                    "/auth/login-with-password",
                    {
                        username: request.data.email,
                        password: request.data.password,
                        trustedDeviceId: request.data.isTrustDevice ? await this.device.getFingerprint() : undefined,
                    },
                    // {
                    //     withCredentials: true,
                    // },
                ),
            ).pipe(
                map(this.validator.signIn),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Send 2-factor authentication token
     */
    async send2FAToken(request: Auth_Send2FAToken_Req): Promise<Result<Auth_Send2FAToken_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.post("/auth/send-mfa-passcode", {
                    mfaToken: request.data.mfaToken,
                }),
            ).pipe(
                map(this.validator.send2FAToken),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Sign in with 2-factor authentication
     */
    async signIn2FA(request: Auth_SignIn2FA_Req): Promise<Result<Auth_SignIn2FA_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.post(
                    "/auth/login-with-passcode",
                    {
                        passcode: request.data.code,
                        mfaToken: request.data.mfaToken,
                    },
                    // {
                    //     withCredentials: true,
                    // },
                ),
            ).pipe(
                map(this.validator.signIn2FA),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Forgot password
     */
    async authorizeDevice(request: Auth_AuthorizeDevice_Req): Promise<Result<Auth_AuthorizeDevice_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.post("/devices/authorize", request.data)).pipe(
                map(this.validator.authorizeDevice),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async forgotPassword(request: Auth_ForgotPassword_Req): Promise<Result<Auth_ForgotPassword_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.post("/users/current/password/forgot", {
                    email: request.data.email,
                }),
            ).pipe(
                map(this.validator.forgotPassword),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Validate reset token
     */
    async validateResetToken(
        request: Auth_ValidateResetToken_Req,
    ): Promise<Result<Auth_ValidateResetToken_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.post("/users/password-reset-token-validate", {
                    email: request.data.email,
                    token: request.data.token,
                }),
            ).pipe(
                map(this.validator.validateResetToken),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Reset password
     */
    async resetPassword(request: Auth_ResetPassword_Req): Promise<Result<Auth_ResetPassword_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.post(`/users/${request.data.userId}/password/reset`, {
                    token: request.data.token,
                    password: request.data.newPassword,
                }),
            ).pipe(
                map(this.validator.resetPassword),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Get login options
     */
    async getLoginOptions(signal: AbortSignal): Promise<Result<Auth_GetLoginOptions_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/auth/login-options", { signal })).pipe(
                map(this.validator.getLoginOptions),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Get sign-up options: whether a person may create their own account
     */
    async getSignupOptions(signal: AbortSignal): Promise<Result<Auth_GetSignupOptions_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/auth/signup-options", { signal })).pipe(
                map(this.validator.getSignupOptions),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * The account a sign-in with a provider offers
     */
    async getSsoSignupInfo(
        request: Auth_GetSsoSignupInfo_Req,
        signal: AbortSignal,
    ): Promise<Result<Auth_GetSsoSignupInfo_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.post("/auth/sso/signup-info", { token: request.data.token }, { signal })).pipe(
                map(this.validator.getSsoSignupInfo),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Make the account a sign-in with a provider offered, and sign it in
     */
    async completeSsoSignup(request: Auth_CompleteSsoSignup_Req): Promise<Result<Auth_CompleteSsoSignup_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.post("/auth/sso/signup", { token: request.data.token })).pipe(
                map(this.validator.completeSsoSignup),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Validate invite token
     */
    async validateInviteToken(
        request: Auth_ValidateInviteToken_Req,
    ): Promise<Result<Auth_ValidateInviteToken_Res, Error>> {
        const { data } = request;

        return lastValueFrom(
            from(
                this.client.v1.post("/users/signup-begin", {
                    inviteToken: data.inviteToken,
                }),
            ).pipe(
                map(this.validator.validateInviteToken),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
