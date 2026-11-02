import { Laptop, ShieldAlert } from "lucide-react";
import { useSearchParams } from "react-router";

import { LogoIcon } from "@assets/icons";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";

import { AuthCommands } from "@application/authentication/data/commands";
import { AuthenticationLayout } from "@application/authentication/layouts";

import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui";

/**
 * An app on somebody's device asks to sign in. Route protection has them signed
 * in by now, however this server signs people in, and `next` brought them back.
 * Nothing happens without the click: a link alone must not sign in a device.
 */
export function ConnectDeviceRoute() {
    const [params] = useSearchParams();

    const { profile } = useProfileContext();

    const { mutate: authorizeDevice, isPending } = AuthCommands.useAuthorizeDevice({
        onSuccess: response => {
            window.location.assign(response.data.redirectUrl);
        },
    });

    const redirectUri = params.get("redirect_uri") ?? "";
    const codeChallenge = params.get("code_challenge") ?? "";
    const state = params.get("state") ?? "";
    const deviceName = params.get("device_name") ?? "";
    const platform = params.get("platform") ?? "";

    const complete = redirectUri !== "" && codeChallenge !== "" && state !== "" && deviceName !== "";

    return (
        <AuthenticationLayout>
            <div className="flex flex-col gap-6">
                <CardHeader className="text-center">
                    <div className="mx-auto mb-2">
                        <LogoIcon className="h-10 w-10" />
                    </div>
                    <CardDescription className="font-medium">Sign in to the shellwing app</CardDescription>
                </CardHeader>
                <div className="w-full max-w-sm mx-auto">
                    <Card>
                        {complete ? (
                            <CardContent className="flex flex-col items-center gap-4 text-center">
                                <div className="rounded-full bg-muted p-3 text-foreground">
                                    <Laptop className="h-6 w-6" />
                                </div>
                                <CardTitle className="text-base">Allow this device to sync?</CardTitle>
                                <CardDescription>
                                    {`${deviceName}${platform !== "" ? ` (${platform})` : ""} will sync as ${
                                        profile?.email ?? profile?.username ?? "you"
                                    }.`}
                                </CardDescription>
                                <CardDescription>
                                    Allow it only if you just started this from the app. Your data stays encrypted: this
                                    gives no one your keys.
                                </CardDescription>
                                <Button
                                    type="button"
                                    className="w-full"
                                    isLoading={isPending}
                                    onClick={() => {
                                        authorizeDevice({ redirectUri, codeChallenge, state, deviceName, platform });
                                    }}
                                >
                                    Allow
                                </Button>
                                <AppLink.Basic
                                    to={ROUTE.home.$route}
                                    className="text-sm text-link underline"
                                >
                                    Not now
                                </AppLink.Basic>
                            </CardContent>
                        ) : (
                            <CardContent className="flex flex-col items-center gap-4 text-center">
                                <div className="rounded-full bg-muted p-3 text-foreground">
                                    <ShieldAlert className="h-6 w-6" />
                                </div>
                                <CardTitle className="text-base">This link is not complete</CardTitle>
                                <CardDescription>Start signing in from the shellwing app again.</CardDescription>
                            </CardContent>
                        )}
                    </Card>
                </div>
            </div>
        </AuthenticationLayout>
    );
}
