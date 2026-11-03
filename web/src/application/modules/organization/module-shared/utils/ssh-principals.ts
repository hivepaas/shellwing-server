/**
 * The names in what was typed: split at commas and white space, without repeats, in the order given.
 */
export function parseSshPrincipals(text: string): string[] {
    const names = text
        .split(/[\s,]+/)
        .map(name => name.trim())
        .filter(name => name.length > 0);

    return [...new Set(names)];
}

/**
 * Names as they are typed in: one line, separated by commas.
 */
export function formatSshPrincipals(names: readonly string[]): string {
    return names.join(", ");
}

/**
 * A lifetime in minutes, as people say it: "8 hours", "1 hour 30 minutes", "45 minutes".
 */
export function formatMinutes(minutes: number): string {
    const hours = Math.floor(minutes / 60);
    const rest = minutes % 60;
    const parts: string[] = [];

    if (hours > 0) {
        parts.push(`${hours} ${hours === 1 ? "hour" : "hours"}`);
    }

    if (rest > 0 || hours === 0) {
        parts.push(`${rest} ${rest === 1 ? "minute" : "minutes"}`);
    }

    return parts.join(" ");
}
