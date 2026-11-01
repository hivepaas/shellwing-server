import { GiteaIcon, GithubIcon, GitlabIcon, GoogleIcon } from "@/assets/icons";
import { DeviceInfo } from "@infrastructure/device";
import { useSearchParams } from "react-router";

import { getSafeNextPathOf } from "@application/shared/utils";

import { type LoginOption } from "@application/authentication/domain";

import { Button } from "@/components/ui/button";

/**
 * The providers one signs in with, or creates an account with: one whose email
 * has no account is offered one to confirm. Where the person was going (`next`)
 * comes along, through the provider and back.
 */
export function SsoButtons({ loginOptions, trustDevice = false }: Props) {
    const [params] = useSearchParams();

    async function signInWith(option: LoginOption) {
        const url = new URL(option.authURL, window.location.origin);
        const next = getSafeNextPathOf(params);
        if (next !== null) {
            url.searchParams.set("next", next);
        }
        // A trusted device is not asked for the passcode again for a while,
        // after a provider's sign-in as after a password.
        const deviceId = trustDevice ? await new DeviceInfo().getFingerprint() : undefined;
        if (deviceId) {
            url.searchParams.set("trustedDeviceId", deviceId);
        }
        window.location.href = url.toString();
    }

    return (
        <>
            {loginOptions.map(option => (
                <Button
                    key={option.authURL}
                    variant="outline"
                    type="button"
                    onClick={() => {
                        void signInWith(option);
                    }}
                >
                    {option.icon === "github" && <GithubIcon />}
                    {option.icon === "gitlab" && <GitlabIcon />}
                    {option.icon === "gitea" && <GiteaIcon />}
                    {option.icon === "google" && <GoogleIcon />}
                    Continue with {option.name}
                </Button>
            ))}
        </>
    );
}

interface Props {
    loginOptions: LoginOption[];
    /** The sign-in page's "Trust this device". */
    trustDevice?: boolean;
}
