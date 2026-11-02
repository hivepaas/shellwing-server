export interface AuthorizeDevice {
    redirectUri: string;
    codeChallenge: string;
    state: string;
    deviceName: string;
    platform: string;
}

export interface ForgotPassword {
    email: string;
}
