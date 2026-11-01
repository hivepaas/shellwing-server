export const SupportFeedbackCategory = {
    General: "general",
    SecurityReport: "security_report",
    BugIssueReport: "bug_issue_report",
} as const;

export type SupportFeedbackCategory = (typeof SupportFeedbackCategory)[keyof typeof SupportFeedbackCategory];
