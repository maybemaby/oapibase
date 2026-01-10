package auth

type SessionUserIdContextKey string
type SessionRoleContextKey string

var SessionUserIdKey SessionUserIdContextKey = "userid"
var SessionRoleKey SessionRoleContextKey = "role"

type SessionData struct {
	UserId int
	Role   string
}
