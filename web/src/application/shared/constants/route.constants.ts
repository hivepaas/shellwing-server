export const ROUTE = {
    /**
     * Authentication
     */
    auth: {
        sso: {
            success: {
                $pattern: "auth/sso/success",
                $route: "/auth/sso/success/",
            },
            // Somebody a provider's sign-in found no account for confirms the one offered.
            signUp: {
                $pattern: "auth/sso/sign-up",
                $route: "/auth/sso/sign-up/",
            },
        },

        twoFA: {
            $pattern: "auth/2fa",
            $route: "/auth/2fa/",
        },

        signIn: {
            $pattern: "auth/sign-in",
            $route: "/auth/sign-in/",
        },

        signUp: {
            $pattern: "auth/sign-up",
            $route: "/auth/sign-up/",
        },
        createAccount: {
            $pattern: "auth/create-account",
            $route: "/auth/create-account/",
        },
        resetPassword: {
            $pattern: "auth/reset-password",
            $route: "/auth/reset-password/",
        },
        forgotPassword: {
            $pattern: "auth/forgot-password",
            $route: "/auth/forgot-password/",
        },
    },

    /**
     * Users
     */
    userManagement: {
        $pattern: "user-management",

        users: {
            $pattern: "user-management/users",
            $route: "/user-management/users/",

            single: {
                $pattern: "user-management/users/:id",
                $route: (id: string) => `/user-management/users/${id}/`,
            },
        },
    },

    /**
     * Current User
     */
    currentUser: {
        $pattern: "current-user",

        profile: {
            $pattern: "current-user/profile",
            $route: "/current-user/profile/",
        },
    },

    /**
     * An app on a device asks to sign in: signed in, and not under auth/, since
     * `next` never leads there.
     */
    connectDevice: {
        $pattern: "connect-device",
        $route: "/connect-device/",
    },

    /**
     * Home: where a signed-in user lands
     */
    home: {
        $pattern: "home",
        $route: "/home/",
    },

    /**
     * Settings
     */
    settings: {
        $pattern: "integrations",

        emailAccounts: {
            $pattern: "integrations/email-accounts",
            $route: "/integrations/email-accounts/",

            create: {
                $pattern: "integrations/email-accounts/create",
                $route: "/integrations/email-accounts/create/",
            },

            edit: {
                $pattern: "integrations/email-accounts/:emailAccountId/edit",
                $route: (emailAccountId: string) => `/integrations/email-accounts/${emailAccountId}/edit/`,
            },
        },

        imPlatforms: {
            $pattern: "integrations/im-platforms",
            $route: "/integrations/im-platforms/",

            create: {
                $pattern: "integrations/im-platforms/create",
                $route: "/integrations/im-platforms/create/",
            },

            edit: {
                $pattern: "integrations/im-platforms/:imPlatformId/edit",
                $route: (imPlatformId: string) => `/integrations/im-platforms/${imPlatformId}/edit/`,
            },
        },

        cloudStorages: {
            $pattern: "integrations/cloud-storages",
            $route: "/integrations/cloud-storages/",

            create: {
                $pattern: "integrations/cloud-storages/create",
                $route: "/integrations/cloud-storages/create/",
            },

            edit: {
                $pattern: "integrations/cloud-storages/:cloudStorageId/edit",
                $route: (cloudStorageId: string) => `/integrations/cloud-storages/${cloudStorageId}/edit/`,
            },
        },

        oauth: {
            $pattern: "integrations/oauth",
            $route: "/integrations/oauth/",

            create: {
                $pattern: "integrations/oauth/create",
                $route: "/integrations/oauth/create/",
            },

            edit: {
                $pattern: "integrations/oauth/:oauthId/edit",
                $route: (oauthId: string) => `/integrations/oauth/${oauthId}/edit/`,
            },
        },

        keyAuth: {
            $pattern: "integrations/key-auth",
            $route: "/integrations/key-auth/",

            create: {
                $pattern: "integrations/key-auth/create",
                $route: "/integrations/key-auth/create/",
            },

            edit: {
                $pattern: "integrations/key-auth/:keyAuthId/edit",
                $route: (keyAuthId: string) => `/integrations/key-auth/${keyAuthId}/edit/`,
            },
        },

        notificationTargets: {
            $pattern: "integrations/notification-targets",
            $route: "/integrations/notification-targets/",

            create: {
                $pattern: "integrations/notification-targets/create",
                $route: "/integrations/notification-targets/create/",
            },

            edit: {
                $pattern: "integrations/notification-targets/:notificationTargetId/edit",
                $route: (notificationTargetId: string) =>
                    `/integrations/notification-targets/${notificationTargetId}/edit/`,
            },
        },
    },

    /**
     * Operations
     */
    operations: {
        $pattern: "operations",
        $route: "/operations/tasks/",

        tasks: {
            $pattern: "operations/tasks",
            $route: "/operations/tasks/",

            details: {
                $pattern: "operations/tasks/:taskId",
                $route: (taskId: string) => `/operations/tasks/${taskId}/`,
            },
        },

        auditLogs: {
            $pattern: "operations/audit-logs",
            $route: "/operations/audit-logs/",
        },
    },
} as const;
