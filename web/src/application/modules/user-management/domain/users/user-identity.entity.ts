/**
 * The public halves of the identity a person's shellwing app made for them, and the fingerprint people compare
 */
export interface UserIdentity {
    /** Base64 */
    x25519Public: string;
    /** Base64 */
    ed25519Public: string;
    /** The X25519 key signed by the Ed25519 one, base64 */
    signature: string;
    /** Six groups of five digits */
    fingerprint: string;
    updatedAt: Date;
}
