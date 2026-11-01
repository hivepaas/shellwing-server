/**
 * The account a sign-in with a provider offers somebody with no account yet,
 * for them to confirm: for this email, under this name.
 */
export interface SsoSignupInfo {
    email: string;
    fullName: string;
}
