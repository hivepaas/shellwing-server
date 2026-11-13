/**
 * What an account name (a principal) may be: the server's own rule. It leaves out what would change meaning where
 * principals are written down - commas split them in a certificate and in sshd's configuration, spaces and # in its
 * files.
 */
export const SSH_PRINCIPAL_PATTERN = /^[A-Za-z0-9_][A-Za-z0-9_.@-]{0,63}$/;

/** How many accounts one member may be granted. */
export const SSH_MAX_PRINCIPALS = 32;

/** How long a certificate may last, in minutes: 5 minutes to a day, 2 hours unless set. */
export const SSH_CERT_TTL_MIN_MINUTES = 5;
export const SSH_CERT_TTL_MAX_MINUTES = 1440;
export const SSH_CERT_TTL_DEFAULT_MINUTES = 120;

/**
 * What a certificate may be allowed to do, by sshd's extension names, with what each one means. An extension absent
 * from a certificate is refused however the target server is configured, so this is the narrower of the two gates.
 *
 * The order is the server's own (pkg/sshca.AllExtensions), which is also the order it stores them in.
 */
export const SSH_CERT_EXTENSIONS = [
    {
        name: "permit-pty",
        label: "Terminal",
        description: "An interactive shell needs one. Without it a certificate can log in and run a command.",
    },
    {
        name: "permit-port-forwarding",
        label: "Port forwarding",
        description:
            "-L, -R and -D. It is how somebody reaches a machine they were never granted, through one they were.",
    },
    {
        name: "permit-agent-forwarding",
        label: "Agent forwarding",
        description:
            "Lets the server use the keys in the person's agent - and so does anybody with root on that server, " +
            "while the session lasts.",
    },
    {
        name: "permit-X11-forwarding",
        label: "X11 forwarding",
        description: "Carries input and the screen both ways.",
    },
    {
        name: "permit-user-rc",
        label: "Run ~/.ssh/rc",
        description: "Runs the target account's own script on each login.",
    },
] as const;

/** What an organization gets until it says otherwise: a terminal, and nothing that reaches past the server. */
export const SSH_CERT_EXTENSIONS_DEFAULT: readonly string[] = ["permit-pty"];

/**
 * Where a target server keeps the authority keys it trusts, as the guide names it.
 *
 * Plural, and fetched rather than copied: an organization may have more than one authority trusted at a time, which is
 * what rotating one needs. A new authority is published before anything is signed with it, so a server that keeps this
 * file up to date has the key before any certificate carries it.
 */
export const SSH_CA_PUBLIC_KEY_PATH = "/etc/ssh/shellwing_ca_keys.pub";

/** The line for a target server's sshd_config. */
export const SSHD_TRUSTED_USER_CA_KEYS_LINE = `TrustedUserCAKeys ${SSH_CA_PUBLIC_KEY_PATH}`;

/** Where a target server keeps the revocation list, as the guide names it. */
export const SSH_KRL_PATH = "/etc/ssh/shellwing_krl";

/** The line that makes a server refuse a certificate that has been revoked. */
export const SSHD_REVOKED_KEYS_LINE = `RevokedKeys ${SSH_KRL_PATH}`;

/** Where the organization's revocation list is fetched from. */
export function sshKrlApiPath(orgId: string): string {
    return `/api/orgs/${orgId}/ssh/krl`;
}

/** Where the authority keys a server should trust are fetched from, with the same token as the list. */
export function sshCaKeysApiPath(orgId: string): string {
    return `/api/orgs/${orgId}/ssh/ca-keys`;
}

/** The folder in the repository that holds the fetch script and its systemd units. */
export const SSH_FETCH_DEPLOYMENT_PATH = "deployment/ssh/";

/**
 * How many authorities an organization may have trusted at once.
 *
 * Room for a rotation, and for one left half done. The server is what enforces it (ERR_SSH_CA_TOO_MANY_TRUSTED); this
 * is only so the button is not offered when it would be refused.
 */
export const SSH_MAX_TRUSTED_CAS = 3;
