package v1

type AuthedRequest struct {
	Username string `header:"X-User-Name" valid:"required"`
}
