package auditserviceimpl

import (
	"context"
	"encoding/json"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/auditsink"
	"github.com/hivepaas/shellwing-server/pkg/logging"
	"github.com/hivepaas/shellwing-server/pkg/reqinfo"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/auditservice"
)

type service struct {
	auditLogRepo repository.AuditLogRepo
	sink         auditsink.Sink
	logger       logging.Logger
}

func New(
	auditLogRepo repository.AuditLogRepo,
	sink auditsink.Sink,
	logger logging.Logger,
) auditservice.Service {
	return &service{
		auditLogRepo: auditLogRepo,
		sink:         sink,
		logger:       logger,
	}
}

// Record implements auditservice.Service.
func (s *service) Record(ctx context.Context, db database.IDB, entry *auditservice.Entry) error {
	if entry == nil {
		return hperrors.NewArgumentInvalid("audit entry")
	}

	auditLog := &entity.AuditLog{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     entry.Scope,
		ObjectID:  entry.ObjectID,
		Type:      entry.Type,
		Source:    entry.Source,
		Section:   entry.Section,
		Result:    entry.Result,
		ResType:   entry.ResType,
		ResID:     entry.ResID,
		ResName:   entry.ResName,
		Detail:    entry.Detail,
		CreatedAt: timeutil.NowUTC(),
	}
	applyActor(auditLog, entry.Auth)
	applyRequest(auditLog, ctx)

	if err := s.auditLogRepo.Insert(ctx, db, auditLog); err != nil {
		return hperrors.Wrap(err)
	}
	return s.send(ctx, auditLog)
}

// send delivers a copy of the record to the sink, if there is one.
//
// After the database and before the caller's transaction commits, which is the only order with a defensible failure.
// A sink holding a record of something that was then rolled back is a false entry; an action that happened with no
// record anywhere is the thing records exist to prevent. Having neither would need the sink inside the transaction,
// and syslog has nothing to commit with - so the false entry is the side to err on, and it is rare: a rollback after
// this point is a failure of something that had already succeeded.
func (s *service) send(ctx context.Context, auditLog *entity.AuditLog) error {
	payload, err := json.Marshal(auditLog)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if err = s.sink.Send(ctx, string(auditLog.Type), payload); err == nil {
		return nil
	}

	if cfg := config.Current(); cfg != nil && !cfg.Audit.Required {
		// Said loudly rather than swallowed. The operator asked for the server to keep working, not for the loss to
		// be invisible, and this line is the only trace that record ever existed outside the database.
		s.logger.Errorf("audit record %s (%s) did not reach the %s sink, and is only in the database: %v",
			auditLog.ID, auditLog.Type, s.sink.Name(), err)
		return nil
	}
	return hperrors.Wrap(hperrors.ErrAuditSinkUnavailable).WithCause(err).
		WithParam("Name", s.sink.Name())
}

// applyActor copies who acted. Every step is guarded because an audit write must
// not be the thing that panics: it runs on paths that are already refusing a
// request, where the auth may be less complete than the happy path assumes.
func applyActor(auditLog *entity.AuditLog, auth *basedto.Auth) {
	auditLog.ActorType = base.SubjectTypeUser
	if auth == nil || auth.User == nil {
		return
	}
	if user := auth.User.Entity(); user != nil {
		auditLog.ActorID = user.ID
		auditLog.ActorName = user.Username
	}
	if claims := auth.User.AuthClaims; claims != nil {
		auditLog.SessionUID = claims.UID
		if auditLog.ActorID == "" {
			auditLog.ActorID = claims.UserID
		}
	}
}

// applyRequest copies where the action came from. Work started outside a request
// - a worker, a scheduled job - carries none of this, and that is not an error.
func applyRequest(auditLog *entity.AuditLog, ctx context.Context) {
	info := reqinfo.From(ctx)
	if info == nil {
		return
	}
	auditLog.RequestID = info.RequestID
	auditLog.ClientIP = info.ClientIP
	auditLog.RemoteAddr = info.RemoteAddr
	auditLog.UserAgent = info.UserAgent
}
