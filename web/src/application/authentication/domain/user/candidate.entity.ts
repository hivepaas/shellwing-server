import { type ESecuritySettings, type EUserRole } from "@application/shared/enums";

/**
 * The person an invitation or a sign-up link is for. `fullName` is the name
 * they were invited or signed up with, to prefill the form that finishes it.
 */
export type Candidate =
    | {
          role: EUserRole;
          email: string;
          fullName: string;
          securityOption: typeof ESecuritySettings.PasswordOnly;
          accessExpiration: Date | null;
      }
    | {
          role: EUserRole;
          email: string;
          fullName: string;
          securityOption: typeof ESecuritySettings.EnforceSSO;
          accessExpiration: Date | null;
      }
    | {
          role: EUserRole;
          email: string;
          fullName: string;
          securityOption: typeof ESecuritySettings.Password2FA;
          accessExpiration: Date | null;
          mfaTotpSecret: string;
          qrCode: string;
      };
