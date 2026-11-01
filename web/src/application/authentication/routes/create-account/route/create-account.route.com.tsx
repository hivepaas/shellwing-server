import { useState } from "react";

import { MailCheck, UserX } from "lucide-react";
import { useSearchParams } from "react-router";

import { LogoIcon } from "@assets/icons";

import { getSafeNextPathOf } from "@application/shared/utils";

import { BackToSignIn } from "@application/authentication/components";
import { AuthCommands } from "@application/authentication/data/commands";
import { AuthQueries } from "@application/authentication/data/queries";
import { AuthenticationLayout } from "@application/authentication/layouts";

import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui";

import { CreateAccountForm } from "../form";
import { type CreateAccountSchemaOutput } from "../schemas";

/**
 * What follows a request, whoever the address belongs to. The server answers
 * the same for an address with an account and one without, and so does this:
 * telling them apart would tell anyone which addresses have accounts.
 */
function SentView({ email, onRestart }: SentViewProps) {
    return (
        <Card>
            <CardContent className="flex flex-col items-center gap-4 text-center">
                <div className="rounded-full bg-muted p-3 text-foreground">
                    <MailCheck className="h-6 w-6" />
                </div>
                <CardTitle className="text-base">Check your email</CardTitle>
                <CardDescription>
                    If <span className="font-medium text-foreground break-all">{email}</span> can be used for a new
                    account, an email with a link to finish creating it is on its way.
                </CardDescription>
                <CardDescription>
                    Nothing after a few minutes? Check your spam folder, or try again. If you already have an account,
                    sign in instead.
                </CardDescription>
                <Button
                    variant="outline"
                    type="button"
                    onClick={onRestart}
                >
                    Try again
                </Button>
                <BackToSignIn />
            </CardContent>
        </Card>
    );
}

interface SentViewProps {
    email: string;
    onRestart: () => void;
}

/**
 * When the server makes accounts by invitation only. The server would refuse
 * the request anyway; this says so before anything is typed.
 */
function ClosedView() {
    return (
        <Card>
            <CardContent className="flex flex-col items-center gap-4 text-center">
                <div className="rounded-full bg-muted p-3 text-foreground">
                    <UserX className="h-6 w-6" />
                </div>
                <CardTitle className="text-base">Accounts are made by invitation</CardTitle>
                <CardDescription>
                    this server does not let you create an account yourself. Ask its administrator for an invitation, or
                    sign in if you already have an account.
                </CardDescription>
                <BackToSignIn />
            </CardContent>
        </Card>
    );
}

export function CreateAccountRoute() {
    const [sentTo, setSentTo] = useState<string | null>(null);

    const [params] = useSearchParams();

    const { data: signupOptions } = AuthQueries.useGetSignupOptions();
    const { data: { data: loginOptions } = { data: [] } } = AuthQueries.useGetLoginOptions();
    const isClosed = signupOptions?.data.selfSignupEnabled === false;

    const { mutate: requestSignUp, isPending } = AuthCommands.useRequestSignUp({
        onSuccess: (_, { email }) => {
            setSentTo(email);
        },
    });

    function handleSubmit(values: CreateAccountSchemaOutput) {
        requestSignUp({
            ...values,
            // Where they were going comes back
            // in the emailed link, and on through signing up and in.
            next: getSafeNextPathOf(params) ?? undefined,
        });
    }

    return (
        <AuthenticationLayout>
            <div className="flex flex-col gap-6">
                <CardHeader className="flex flex-col items-center text-center">
                    <LogoIcon className="h-12 w-12 text-foreground" />
                    <CardDescription className="font-medium">Create your shellwing account</CardDescription>
                </CardHeader>
                <div className="w-full max-w-sm mx-auto">
                    {isClosed ? (
                        <ClosedView />
                    ) : sentTo === null ? (
                        <CreateAccountForm
                            loginOptions={loginOptions}
                            isPending={isPending}
                            onSubmit={handleSubmit}
                        />
                    ) : (
                        <SentView
                            email={sentTo}
                            onRestart={() => {
                                setSentTo(null);
                            }}
                        />
                    )}
                </div>
            </div>
        </AuthenticationLayout>
    );
}
