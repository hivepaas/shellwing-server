/**
 * Asking for an account: the server emails a link to finish creating it.
 */
export interface SignUpRequest {
    email: string;
    fullName: string;
    /** Where to go once the account is made and signed in to: a path on this origin. */
    next?: string;
}
