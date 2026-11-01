import { UserPlus, UserX } from "lucide-react";
import { useCookie } from "react-use";

import { LogoIcon } from "@assets/icons";

import { AppLoader } from "@application/shared/components";
import { useProfileContext } from "@application/shared/context";

import { BackToSignIn } from "@application/authentication/components";
import { AuthCommands } from "@application/authentication/data/commands";
import { AuthQueries } from "@application/authentication/data/queries";
import { AuthenticationLayout } from "@application/authentication/layouts";

import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui";

/** The cookie a provider's sign-in hands over for somebody with no account. */
const SSO_SIGNUP_TOKEN_COOKIE = "sso_signup_token";

function UnavailableView() {
    return (
        <Card>
            <CardContent className="flex flex-col items-center gap-4 text-center">
                <div className="rounded-full bg-muted p-3 text-foreground">
                    <UserX className="h-6 w-6" />
                </div>
                <CardTitle className="text-base">This sign-up has run out</CardTitle>
                <CardDescription>
                    The account offered after signing in with your provider can no longer be created from here. Sign in
                    with the provider again, or create an account with your email.
                </CardDescription>
                <BackToSignIn />
            </CardContent>
        </Card>
    );
}

/**
 * A sign-in with a provider found no account for its email: this asks before
 * one is made, so that somebody whose account has another email does not end
 * up with a second, empty one.
 */
export function SsoSignUpRoute() {
    const [token, , deleteToken] = useCookie(SSO_SIGNUP_TOKEN_COOKIE);

    const { setProfile } = useProfileContext();

    const { data: info, isPending: isLoading, isError } = AuthQueries.useGetSsoSignupInfo(token ?? "");

    const { mutate: completeSsoSignup, isPending } = AuthCommands.useCompleteSsoSignup({
        onSuccess: profile => {
            deleteToken();

            setProfile(profile);
        },
    });

    let child;
    if (!token || isError) {
        child = <UnavailableView />;
    } else if (isLoading) {
        return <AppLoader />;
    } else {
        const { email, fullName } = info.data;

        child = (
            <Card>
                <CardContent className="flex flex-col items-center gap-4 text-center">
                    <div className="rounded-full bg-muted p-3 text-foreground">
                        <UserPlus className="h-6 w-6" />
                    </div>
                    <CardTitle className="text-base">Create your shellwing account?</CardTitle>
                    <CardDescription>
                        No account has <span className="font-medium text-foreground break-all">{email}</span> yet.
                        {fullName !== "" && <> Create one for {fullName}, signing in with your provider.</>}
                    </CardDescription>
                    <CardDescription>
                        Already have an account under another email? Go back and sign in with it instead.
                    </CardDescription>
                    <Button
                        type="button"
                        className="w-full"
                        isLoading={isPending}
                        onClick={() => {
                            completeSsoSignup({ token });
                        }}
                    >
                        Create account
                    </Button>
                    <BackToSignIn />
                </CardContent>
            </Card>
        );
    }

    return (
        <AuthenticationLayout>
            <div className="flex flex-col gap-6">
                <CardHeader className="flex flex-col items-center text-center">
                    <LogoIcon className="h-12 w-12 text-foreground" />
                    <CardDescription className="font-medium">Create your shellwing account</CardDescription>
                </CardHeader>
                <div className="w-full max-w-sm mx-auto">{child}</div>
            </div>
        </AuthenticationLayout>
    );
}
