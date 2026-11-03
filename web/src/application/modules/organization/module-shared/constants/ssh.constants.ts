/**
 * What an account name (a principal) may be: the server's own rule. It leaves out what would change meaning where
 * principals are written down - commas split them in a certificate and in sshd's configuration, spaces and # in its
 * files.
 */
export const SSH_PRINCIPAL_PATTERN = /^[A-Za-z0-9_][A-Za-z0-9_.@-]{0,63}$/;

/** How many accounts one member may be granted. */
export const SSH_MAX_PRINCIPALS = 32;

/** How long a certificate may last, in minutes: 5 minutes to a day, 8 hours unless set. */
export const SSH_CERT_TTL_MIN_MINUTES = 5;
export const SSH_CERT_TTL_MAX_MINUTES = 1440;
export const SSH_CERT_TTL_DEFAULT_MINUTES = 480;

/** Where a target server keeps the CA's public key, as the guide names it. */
export const SSH_CA_PUBLIC_KEY_PATH = "/etc/ssh/shellwing_ca.pub";

/** The line for a target server's sshd_config. */
export const SSHD_TRUSTED_USER_CA_KEYS_LINE = `TrustedUserCAKeys ${SSH_CA_PUBLIC_KEY_PATH}`;
