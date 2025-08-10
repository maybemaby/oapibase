package auth

const SessionUserIdKey = "userid"
const SessionRoleKey = "role"

type SessionData struct {
	UserId int
	Role   string
}
