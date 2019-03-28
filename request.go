package slash

import "net/url"

// Request is the request message for a slash command.
type Request url.Values

func (r Request) get(field string) string { return url.Values(r).Get(field) }

// Token returns the `token` field from the slash request.
func (r Request) Token() string { return r.get("token") }

// TeamID returns the `team_id` field from the slash request.
func (r Request) TeamID() string { return r.get("team_id") }

// TeamDomain returns the `team_domain` field from the slash request.
func (r Request) TeamDomain() string { return r.get("team_domain") }

// EnterpriseID returns the `enterprise_id` field from the slash request.
func (r Request) EnterpriseID() string { return r.get("enterprise_id") }

// EnterpriseName returns the `enterprise_name` field from the slash request.
func (r Request) EnterpriseName() string { return r.get("enterprise_name") }

// ChannelID returns the `channel_id` field from the slash request.
func (r Request) ChannelID() string { return r.get("channel_id") }

// ChannelName returns the `channel_name` field from the slash request.
func (r Request) ChannelName() string { return r.get("channel_name") }

// UserID returns the `user_id` field from the slash request.
func (r Request) UserID() string { return r.get("user_id") }

// UserName returns the `user_name` field from the slash request.
func (r Request) UserName() string { return r.get("user_name") }

// Command returns the `command` field from the slash request.
func (r Request) Command() string { return r.get("command") }

// Text returns the `text` field from the slash request.
func (r Request) Text() string { return r.get("text") }

// ResponseURL returns the `response_url` field from the slash request.
func (r Request) ResponseURL() string { return r.get("response_url") }

// TriggerID returns the `trigger_id` field from the slash request.
func (r Request) TriggerID() string { return r.get("trigger_id") }
