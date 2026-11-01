import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";

import { BackToSignIn, SsoButtons } from "@application/authentication/components";
import { type LoginOption } from "@application/authentication/domain";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader } from "@/components/ui/card";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";

import { CreateAccountSchema, type CreateAccountSchemaInput, type CreateAccountSchemaOutput } from "../schemas";

export function CreateAccountForm({ loginOptions, isPending, onSubmit }: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<CreateAccountSchemaInput, unknown, CreateAccountSchemaOutput>({
        defaultValues: {
            email: "",
            fullName: "",
        },
        resolver: zodResolver(CreateAccountSchema),
        mode: "onSubmit",
    });

    const {
        field: email,
        fieldState: { invalid: isEmailInvalid },
    } = useController({
        name: "email",
        control,
    });

    const {
        field: fullName,
        fieldState: { invalid: isFullNameInvalid },
    } = useController({
        name: "fullName",
        control,
    });

    function onValid(values: CreateAccountSchemaOutput) {
        void onSubmit(values);
    }

    function onInvalid(_: FieldErrors<CreateAccountSchemaOutput>) {
        //
    }

    return (
        <div className="flex flex-col gap-6">
            <Card>
                <CardHeader>
                    <CardDescription>We will email you a link to finish creating your account.</CardDescription>
                </CardHeader>
                <CardContent>
                    <form
                        onSubmit={event => {
                            event.preventDefault();

                            void handleSubmit(onValid, onInvalid)(event);
                        }}
                    >
                        <FieldGroup>
                            <Field>
                                <FieldLabel
                                    htmlFor="email"
                                    isRequired
                                >
                                    Email
                                </FieldLabel>
                                <Input
                                    id="email"
                                    value={email.value}
                                    onChange={email.onChange}
                                    onBlur={email.onBlur}
                                    type="email"
                                    autoComplete="email"
                                    placeholder="m@example.com"
                                    aria-invalid={isEmailInvalid}
                                    required
                                />
                                <FieldError errors={[errors.email]} />
                            </Field>
                            <Field>
                                <FieldLabel
                                    htmlFor="fullName"
                                    isRequired
                                >
                                    Full Name
                                </FieldLabel>
                                <Input
                                    id="fullName"
                                    value={fullName.value}
                                    onChange={fullName.onChange}
                                    onBlur={fullName.onBlur}
                                    type="text"
                                    autoComplete="name"
                                    placeholder="Enter your full name"
                                    aria-invalid={isFullNameInvalid}
                                    required
                                />
                                <FieldError errors={[errors.fullName]} />
                            </Field>
                            <Field>
                                <Button
                                    type="submit"
                                    isLoading={isPending}
                                >
                                    Create account
                                </Button>
                                {/* Or with a provider: an email it vouches for needs no link to confirm. */}
                                <SsoButtons loginOptions={loginOptions} />
                                <BackToSignIn />
                            </Field>
                        </FieldGroup>
                    </form>
                </CardContent>
            </Card>
        </div>
    );
}

interface Props {
    loginOptions: LoginOption[];
    isPending: boolean;
    onSubmit: (values: CreateAccountSchemaOutput) => Promise<void> | void;
}
