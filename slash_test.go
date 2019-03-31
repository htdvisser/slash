package slash_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"htdvisser.dev/slash"
)

func Example() {
	// This is an extremely simple response to a slash command.
	// You'll want to add "blocks" for more advanced responses.
	// See also https://api.slack.com/messaging/composing/layouts.
	type response struct {
		ResponseType string `json:"response_type"`
		Text         string `json:"text"`
	}

	// We set up the slash command router with the Signing secret we got from Slack.
	r := slash.NewRouter(os.Getenv("SLACK_SIGNING_SECRET"))

	// We register the /pseudorandom command that generates a pseudorandom number for the user.
	r.RegisterCommand("/pseudorandom", func(ctx context.Context, req slash.Request) interface{} {
		n := 10
		if req.Text() != "" {
			var err error
			n, err = strconv.Atoi(req.Text())
			if err != nil {
				// We send a response to the user to inform them about errors.
				return response{
					ResponseType: "ephemeral",
					Text:         "Please give me a number",
				}
			}
		}
		if n <= 1 {
			return response{
				ResponseType: "ephemeral",
				Text:         "Please give me a number _larger than one_",
			}
		}
		return response{
			ResponseType: "in_channel",
			Text:         fmt.Sprintf("%s got %d", req.UserName(), rand.Intn(n)),
		}
	})

	http.Handle("/slash", r)
	http.ListenAndServe(":8080", nil)
}
