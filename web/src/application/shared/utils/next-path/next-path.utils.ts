/**
 * `next` is where someone goes once signed in. It rides along through signing
 * in, creating an account, the email that verifies it and signing up, and since
 * it comes from the address bar, anyone can write it. So it is only ever
 * followed to a path on this origin.
 */
const NEXT_PARAM = "next";

export function normalizePathname(pathname: string): string {
    const withLeadingSlash = pathname.startsWith("/") ? pathname : `/${pathname}`;

    return withLeadingSlash.replace(/\/+$/, "") || "/";
}

function isAuthPath(pathname: string): boolean {
    const normalizedPathname = normalizePathname(pathname);

    return normalizedPathname === "/auth" || normalizedPathname.startsWith("/auth/");
}

export function parseSameOriginUrl(path: string | null): URL | null {
    const trimmedPath = path?.trim();

    if (!trimmedPath) {
        return null;
    }

    try {
        const url = new URL(trimmedPath, `${window.location.origin}/`);

        if (url.origin !== window.location.origin) {
            return null;
        }

        return url;
    } catch {
        return null;
    }
}

/**
 * Whether a path on this origin still reads as one to whatever reads it next:
 * the browser, or the server server, which holds a `next` it mails to the same
 * rules. "/.//example.com" is on this origin, yet parses to "//example.com",
 * which a browser takes for another host; a backslash can read as a slash; and
 * a broken escape fails a stricter parser.
 */
function isPlainPath(pathname: string, path: string): boolean {
    return !pathname.includes("//") && !path.includes("\\") && !/%(?![0-9a-f]{2})/i.test(path);
}

/**
 * The path, query and hash `next` points at, or null when there is none worth
 * following: another origin, an auth page, where signing in would only loop, or
 * a path that is not plain.
 */
export function getSafeNextPath(path: string | null): string | null {
    const url = parseSameOriginUrl(path);

    if (url === null || isAuthPath(url.pathname)) {
        return null;
    }

    const nextPath = `${url.pathname}${url.search}${url.hash}`;

    return isPlainPath(url.pathname, nextPath) ? nextPath : null;
}

/**
 * The `next` a page was given, still safe to follow, if it has one.
 */
export function getSafeNextPathOf(params: URLSearchParams): string | null {
    return getSafeNextPath(params.get(NEXT_PARAM));
}

/**
 * The query that carries `next` on to another page: empty without one.
 */
export function toNextPathSearch(nextPath: string | null): string {
    return nextPath === null ? "" : new URLSearchParams({ [NEXT_PARAM]: nextPath }).toString();
}
