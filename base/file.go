package base

const (
	// 0755 grants read/write/execute for owner, read/execute for group/others
	DirModeDefault = 0755
)

type FileStatus string

const (
	FileStatusActive   FileStatus = "active"
	FileStatusPending  FileStatus = "pending"
	FileStatusDisabled FileStatus = "disabled"
	FileStatusDeleting FileStatus = "deleting"
)

var (
	AllFileStatuses = []FileStatus{FileStatusActive, FileStatusPending, FileStatusDisabled, FileStatusDeleting}
)

type FileType string

const (
	FileTypeDataFile       FileType = "data-file"
	FileTypeSystem         FileType = "system"
	FileTypeCache          FileType = "cache"
	FileTypeTmp            FileType = "tmp"
	FileTypeSchedJobOutput FileType = "sched-job-output"
)

var (
	AllFileTypes = []FileType{FileTypeDataFile, FileTypeSystem, FileTypeCache, FileTypeTmp,
		FileTypeSchedJobOutput}
)

type FileStorageType string

const (
	FileStorageLocal FileStorageType = "local"
	FileStorageCloud FileStorageType = "cloud"
)

var (
	AllFileStorageTypes = []FileStorageType{FileStorageLocal, FileStorageCloud}
)
