package apiutils

type Serveable interface {
	ListenAndServe(adr string) error
}
