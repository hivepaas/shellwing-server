/**
 * How a person may get an account here.
 */
export interface SignupOptions {
    /**
     * Anybody may create one with an email address they verify. Off, an account
     * is only made by an admin's invitation.
     */
    selfSignupEnabled: boolean;
}
