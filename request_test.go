package slash_test

import (
	"fmt"
	"testing"

	"htdvisser.dev/slash"
)

func TestRequest(t *testing.T) {
	req := make(slash.Request)
	tt := map[string]func() string{
		"token":           req.Token,
		"team_id":         req.TeamID,
		"team_domain":     req.TeamDomain,
		"enterprise_id":   req.EnterpriseID,
		"enterprise_name": req.EnterpriseName,
		"channel_id":      req.ChannelID,
		"channel_name":    req.ChannelName,
		"user_id":         req.UserID,
		"user_name":       req.UserName,
		"command":         req.Command,
		"text":            req.Text,
		"response_url":    req.ResponseURL,
		"trigger_id":      req.TriggerID,
	}
	for field := range tt {
		req[field] = []string{fmt.Sprintf("%s value", field)}
	}
	for field, fun := range tt {
		t.Run(field, func(t *testing.T) {
			expected := fmt.Sprintf("%s value", field)
			if got := fun(); got != expected {
				t.Fatalf("Expected %s but got %s", expected, got)
			}
		})
	}

}
