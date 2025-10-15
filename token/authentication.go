package token

type Principal interface {
	GetName() string
}

type Authentication struct {
	UserId   *string `json:"userId"`
	TenantId *int    `json:"tenantId"`
}

func (a *Authentication) GetName() string {
	if a.UserId == nil {
		return ""
	}
	return *a.UserId
}
