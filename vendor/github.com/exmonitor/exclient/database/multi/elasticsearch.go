package multi

import (
	"fmt"
	"time"

	"github.com/exmonitor/exclient/database/spec/status"
)

// **************************************************
// ELASTIC SEARCH
///--------------------------------------------------
func (c *Client) ES_GetFailedServices(from time.Time, to time.Time, interval int) ([]*status.ServiceStatus, error) {
	// just dummy record return
	fmt.Printf("ES_GetFailedServices - NOT IMPLEMENTED")

	return nil, nil
}

func (c *Config) ES_SaveServiceStatus(s *status.ServiceStatus) error {
	// TODO
	fmt.Printf("ES_SaveServiceStatus - NOT IMPLEMENTED")
	return nil
}
