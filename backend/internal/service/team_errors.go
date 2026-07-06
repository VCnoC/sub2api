package service

import infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"

var (
	ErrAlreadyInTeam           = infraerrors.Conflict("ALREADY_IN_TEAM", "user is already in a team")
	ErrTeamNotFound            = infraerrors.NotFound("TEAM_NOT_FOUND", "team not found")
	ErrInviteCodeInvalid       = infraerrors.BadRequest("INVITE_CODE_INVALID", "invalid invite code")
	ErrInviteCodeExists        = infraerrors.Conflict("INVITE_CODE_EXISTS", "invite code already exists")
	ErrNotTeamOwner            = infraerrors.Forbidden("NOT_TEAM_OWNER", "only team owner can perform this action")
	ErrNotInTeam               = infraerrors.Forbidden("NOT_IN_TEAM", "user is not in a team")
	ErrCannotRemoveOwner       = infraerrors.Forbidden("CANNOT_REMOVE_OWNER", "cannot remove team owner")
	ErrCannotLeaveAsOwner      = infraerrors.Forbidden("CANNOT_LEAVE_AS_OWNER", "team owner cannot leave the team")
	ErrTeamMemberNotFound      = infraerrors.NotFound("TEAM_MEMBER_NOT_FOUND", "team member not found")
	ErrInsufficientTeamBalance = infraerrors.BadRequest("INSUFFICIENT_TEAM_BALANCE", "insufficient balance for transfer")
	ErrTeamPasswordIncorrect   = infraerrors.BadRequest("TEAM_PASSWORD_INCORRECT", "incorrect password")
)
