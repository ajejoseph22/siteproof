package domain

// Permission represents an action that can be authorized for a role.
type Permission string

const (
	PermissionProjectRead            Permission = "project:read"
	PermissionProjectWrite           Permission = "project:write"
	PermissionTicketRead             Permission = "ticket:read"
	PermissionTicketWrite            Permission = "ticket:write"
	PermissionTicketSubmit           Permission = "ticket:submit"
	PermissionTicketReject           Permission = "ticket:reject"
	PermissionSignatureRequest       Permission = "signature:request"
)

var rolePermissions = map[Role]map[Permission]struct{}{
	RoleForeman: {
		PermissionProjectRead:      {},
		PermissionTicketRead:       {},
		PermissionTicketWrite:      {},
		PermissionTicketSubmit:     {},
		PermissionSignatureRequest: {},
	},
	RoleGCSuper: {
		PermissionProjectRead:  {},
		PermissionTicketRead:   {},
		PermissionTicketReject: {},
	},
	RolePM: {
		PermissionProjectRead:      {},
		PermissionProjectWrite:     {},
		PermissionTicketRead:       {},
		PermissionTicketWrite:      {},
		PermissionTicketReject:     {},
		PermissionSignatureRequest: {},
	},
	RoleAccounting: {
		PermissionProjectRead: {},
		PermissionTicketRead:  {},
	},
	RoleAdmin: {
		PermissionProjectRead:      {},
		PermissionProjectWrite:     {},
		PermissionTicketRead:       {},
		PermissionTicketWrite:      {},
		PermissionTicketSubmit:     {},
		PermissionTicketReject:     {},
		PermissionSignatureRequest: {},
	},
}

// RoleHasPermission reports whether the role can perform the permission.
func RoleHasPermission(role Role, permission Permission) bool {
	perms, ok := rolePermissions[role]
	if !ok {
		return false
	}
	_, ok = perms[permission]
	return ok
}
