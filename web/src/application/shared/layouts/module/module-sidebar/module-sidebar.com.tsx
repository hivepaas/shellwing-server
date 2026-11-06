import * as React from "react";

import { LogoIcon } from "@/assets/icons";
import {
    BadgeCheck,
    Bell,
    CircleDashed,
    CircleUser,
    FileKey,
    FolderLock,
    Group,
    HardDrive,
    IdCard,
    KeySquare,
    ListTodo,
    type LucideIcon,
    Mail,
    MessageSquare,
    PanelLeftClose,
    PanelLeftOpen,
    Puzzle,
    ScrollText,
    ShieldCheck,
    Users,
    UsersRound,
} from "lucide-react";
import { Link, useLocation } from "react-router";

import { MODULE_IDS, ROUTE, type ResourceModuleId } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";
import { useAdministeredOrg } from "@application/shared/hooks";
import { type ModuleId, type ModulePermission, useConditionalModuleCollections } from "@application/shared/permissions";

import { Button, Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui";
import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarHeader,
    SidebarRail,
    useSidebar,
} from "@/components/ui/sidebar";

import { NavMain } from "../nav-main";
import { NavUser } from "../nav-user";

interface SidebarSubItem {
    title: string;
    route: string;
    pattern: string;
    icon?: LucideIcon;
    moduleId?: ResourceModuleId;
    /** Only for an owner or an admin of an organization: their role there, not a module permission, decides. */
    requiresOrgAdmin?: boolean;
}

interface SidebarSection {
    title: string;
    items: SidebarSubItem[];
}

interface SidebarItem {
    title: string;
    route: string;
    pattern: string;
    icon?: LucideIcon;
    moduleId?: ResourceModuleId;
    requiresOrgAdmin?: boolean;
    alwaysVisible?: boolean;
    items?: SidebarSubItem[];
    sections?: SidebarSection[];
}

const navMain: SidebarItem[] = [
    {
        title: "Integrations",
        route: "#",
        pattern: ROUTE.settings.$pattern,
        icon: Puzzle,
        moduleId: MODULE_IDS.Settings,
        sections: [
            {
                title: "Access & Authentication",
                items: [
                    {
                        title: "OAuth",
                        icon: IdCard,
                        route: ROUTE.settings.oauth.$route,
                        pattern: ROUTE.settings.oauth.$pattern,
                    },
                    {
                        title: "Key Auth",
                        icon: KeySquare,
                        route: ROUTE.settings.keyAuth.$route,
                        pattern: ROUTE.settings.keyAuth.$pattern,
                    },
                ],
            },
            {
                title: "Storages",
                items: [
                    {
                        title: "Cloud Storages",
                        icon: HardDrive,
                        route: ROUTE.settings.cloudStorages.$route,
                        pattern: ROUTE.settings.cloudStorages.$pattern,
                    },
                ],
            },
            {
                title: "Notifications",
                items: [
                    {
                        title: "Email Accounts",
                        icon: Mail,
                        route: ROUTE.settings.emailAccounts.$route,
                        pattern: ROUTE.settings.emailAccounts.$pattern,
                    },
                    {
                        title: "IM Platforms",
                        icon: MessageSquare,
                        route: ROUTE.settings.imPlatforms.$route,
                        pattern: ROUTE.settings.imPlatforms.$pattern,
                    },
                    {
                        title: "Notification Targets",
                        icon: Bell,
                        route: ROUTE.settings.notificationTargets.$route,
                        pattern: ROUTE.settings.notificationTargets.$pattern,
                    },
                ],
            },
        ],
    },
    {
        title: "Operations",
        route: "#",
        pattern: ROUTE.operations.$pattern,
        icon: CircleDashed,
        moduleId: MODULE_IDS.System,
        items: [
            {
                title: "Tasks",
                icon: ListTodo,
                route: ROUTE.operations.tasks.$route,
                pattern: ROUTE.operations.tasks.$pattern,
            },
            {
                title: "Audit Logs",
                icon: ScrollText,
                route: ROUTE.operations.auditLogs.$route,
                pattern: ROUTE.operations.auditLogs.$pattern,
            },
        ],
    },
    {
        title: "User Management",
        route: ROUTE.userManagement.users.$route,
        pattern: ROUTE.userManagement.users.$pattern,
        icon: Users,
        moduleId: MODULE_IDS.User,
    },
    {
        title: "Administration",
        route: "#",
        pattern: "#",
        icon: ShieldCheck,
        items: [
            {
                title: "Members",
                icon: UsersRound,
                route: ROUTE.organization.members.$route,
                pattern: ROUTE.organization.members.$pattern,
                requiresOrgAdmin: true,
            },
            {
                title: "Groups",
                icon: Group,
                route: ROUTE.organization.groups.$route,
                pattern: ROUTE.organization.groups.$pattern,
                requiresOrgAdmin: true,
            },
            {
                title: "Collections",
                icon: FolderLock,
                route: ROUTE.organization.collections.$route,
                pattern: ROUTE.organization.collections.$pattern,
                requiresOrgAdmin: true,
            },
            {
                title: "SSH Certificates",
                icon: FileKey,
                route: ROUTE.organization.sshCertificates.$route,
                pattern: ROUTE.organization.sshCertificates.$pattern,
                requiresOrgAdmin: true,
            },
            {
                title: "License",
                icon: BadgeCheck,
                route: ROUTE.system.license.$route,
                pattern: ROUTE.system.license.$pattern,
                moduleId: MODULE_IDS.System,
            },
        ],
    },
    {
        title: "Profile",
        route: ROUTE.currentUser.profile.$route,
        pattern: ROUTE.currentUser.profile.$pattern,
        icon: CircleUser,
        alwaysVisible: true,
    },
];

interface SidebarAccess {
    permissions: ReadonlyMap<ModuleId, ModulePermission>;
    isOrgAdmin: boolean;
}

function hasReadableModuleAccess(
    item: { moduleId?: ResourceModuleId; requiresOrgAdmin?: boolean; alwaysVisible?: boolean },
    { permissions, isOrgAdmin }: SidebarAccess,
) {
    if (item.alwaysVisible ?? false) {
        return true;
    }

    if ((item.requiresOrgAdmin ?? false) && !isOrgAdmin) {
        return false;
    }

    return !item.moduleId || permissions.get(item.moduleId)?.actions.read === true;
}

function filterSidebarItems(items: readonly SidebarItem[], access: SidebarAccess): SidebarItem[] {
    return items.flatMap((item): SidebarItem[] => {
        if (!hasReadableModuleAccess(item, access)) {
            return [];
        }

        if (item.sections && item.sections.length > 0) {
            const filteredSections = item.sections.flatMap(sec => {
                const subItems = sec.items.filter(subItem => hasReadableModuleAccess(subItem, access));

                if (subItems.length === 0) {
                    return [];
                }

                return [
                    {
                        ...sec,
                        items: subItems,
                    },
                ];
            });

            if (filteredSections.length === 0) {
                return [];
            }

            return [
                {
                    ...item,
                    sections: filteredSections,
                },
            ];
        }

        const subItems = item.items?.filter(subItem => hasReadableModuleAccess(subItem, access));

        if (item.items && (!subItems || subItems.length === 0)) {
            return [];
        }

        return [
            {
                ...item,
                items: subItems,
            },
        ];
    });
}

export function ModuleSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
    const { profile } = useProfileContext();
    const { map: modulePermissionMap } = useConditionalModuleCollections();
    // Asked for once someone is signed in: an organization's admin pages are offered by their role there.
    const { org: administeredOrg } = useAdministeredOrg({ enabled: profile !== null });
    const isOrgAdmin = administeredOrg !== null;
    const { isMobile, setOpenMobile, toggleSidebar } = useSidebar();
    const location = useLocation();

    React.useEffect(() => {
        if (isMobile) {
            setOpenMobile(false);
        }
    }, [location.pathname, isMobile, setOpenMobile]);

    const navigationItems = React.useMemo(
        () => filterSidebarItems(navMain, { permissions: modulePermissionMap, isOrgAdmin }),
        [modulePermissionMap, isOrgAdmin],
    );

    if (!profile) {
        return null;
    }
    return (
        <Sidebar
            collapsible="icon"
            {...props}
        >
            <SidebarHeader className="p-2">
                {/* Expanded state: Prominent Logo + shellwing brand + Collapse button */}
                <div className="flex items-center justify-between px-2 py-1 group-data-[collapsible=icon]:hidden">
                    <Link
                        to={ROUTE.home.$route}
                        className="flex items-center gap-2.5 overflow-hidden rounded-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                    >
                        <LogoIcon className="x-logo size-10 shrink-0 text-foreground" />
                        <span className="font-bold text-lg tracking-tight text-foreground truncate">shellwing</span>
                    </Link>
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Button
                                variant="ghost"
                                size="icon"
                                className="size-8 text-muted-foreground hover:text-foreground hover:bg-accent shrink-0"
                                onClick={toggleSidebar}
                                aria-label="Collapse sidebar"
                            >
                                <PanelLeftClose className="size-4" />
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent side="right">Collapse sidebar</TooltipContent>
                    </Tooltip>
                </div>

                {/* Collapsed state: Logo on top + Expand button below */}
                <div className="hidden group-data-[collapsible=icon]:flex flex-col items-center gap-2 py-1">
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Link
                                to={ROUTE.home.$route}
                                className="flex items-center justify-center rounded-md hover:opacity-80 transition-opacity focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                                aria-label="shellwing Home"
                            >
                                <LogoIcon className="x-logo size-9 text-foreground" />
                            </Link>
                        </TooltipTrigger>
                        <TooltipContent side="right">shellwing</TooltipContent>
                    </Tooltip>

                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Button
                                variant="ghost"
                                size="icon"
                                className="size-8 text-muted-foreground hover:text-foreground hover:bg-accent shrink-0"
                                onClick={toggleSidebar}
                                aria-label="Expand sidebar"
                            >
                                <PanelLeftOpen className="size-4" />
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent side="right">Expand sidebar</TooltipContent>
                    </Tooltip>
                </div>
            </SidebarHeader>
            <SidebarContent>
                <NavMain items={navigationItems} />
            </SidebarContent>
            <SidebarFooter>
                <NavUser user={profile} />
            </SidebarFooter>
            <SidebarRail />
        </Sidebar>
    );
}
