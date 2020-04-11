package gatekeeper

type FreePortService struct {
	freePorts map[string]bool
}

func newFreePortService() *FreePortService {
	return &FreePortService{freePorts: make(map[string]bool)}
}

func (r *FreePortService) getFreePort() {

}
