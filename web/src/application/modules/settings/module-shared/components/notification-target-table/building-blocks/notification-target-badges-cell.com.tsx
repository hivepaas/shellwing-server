import { Badge } from "@components/ui/badge";
import type { SettingNotification } from "~/settings/domain";
import { channelTone } from "~/settings/module-shared/constants";

function getTargets(notification: SettingNotification): string[] {
    const targets: string[] = [];

    if (notification.viaEmail?.enabled) {
        targets.push("email");
    }
    if (notification.viaSlack?.enabled) {
        targets.push("slack");
    }
    if (notification.viaDiscord?.enabled) {
        targets.push("discord");
    }

    return targets;
}

export function NotificationTargetBadgesCell({ notification }: { notification: SettingNotification }) {
    const targets = getTargets(notification);

    if (targets.length === 0) {
        return "-";
    }

    return (
        <div className="flex flex-wrap gap-1.5">
            {targets.map(target => (
                <Badge
                    key={target}
                    tone={channelTone(target)}
                >
                    {target}
                </Badge>
            ))}
        </div>
    );
}
