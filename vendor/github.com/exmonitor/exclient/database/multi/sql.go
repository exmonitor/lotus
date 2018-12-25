package multi

import (
	"fmt"
	"github.com/exmonitor/exclient/database/spec/notification"
	"github.com/exmonitor/exclient/database/spec/service"
)

// ********************************************
// MARIA DB
//----------------------------------------------
func (c *Client) SQL_GetIntervals() ([]int, error) {
	// TODO
	fmt.Printf("SQL_GetIntervals - NOT IMPLEMENTED")

	return []int{30, 60, 120, 300}, nil
}

func (c *Client) SQL_GetUsersNotificationSettings(serviceID int) ([]*notification.UserNotificationSettings, error) {

	return nil, nil
}

func (c *Client) SQL_GetServices(interval int) ([]*service.Service, error) {
	var services []*service.Service

	fmt.Printf("SQL_GetServices - NOT IMPLEMENTED")

	return services, nil
}

func (c *Client) SQL_GetServiceDetails(serviceID int) (*service.Service, error) {

	fmt.Printf("SQL_GetServiceDetails - NOT IMPLEMENTED")

	return nil, nil
}
