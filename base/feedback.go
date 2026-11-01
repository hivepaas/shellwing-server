package base

type FeedbackType string

const (
	FeedbackTypeSecurity FeedbackType = "security"
	FeedbackTypeIssue    FeedbackType = "issue"
	FeedbackTypeGeneral  FeedbackType = "general"
)

var (
	AllFeedbackTypes = []FeedbackType{FeedbackTypeSecurity, FeedbackTypeIssue, FeedbackTypeGeneral}
)
