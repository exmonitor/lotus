package status

import "time"

type ServiceStatus struct {
	Id            int
	FailThreshold int
	ResentEvery   time.Duration
	Duration      time.Duration
	Message       string
	Result        bool
	ReqId         string
}

// find status in the status array
func FindStatus(s []*ServiceStatus, id int) *ServiceStatus {
	for _, status := range s {
		if status.Id == id {
			return status
		}
	}
	return nil
}

// check if status with specific id exists in the status array
func Exists(s []*ServiceStatus, id int) bool {
	return FindStatus(s, id) != nil
}
