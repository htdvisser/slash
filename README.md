# Slash

A router for Slack's [slash commands](https://api.slack.com/slash-commands).

## Preparations

1. Go to Slack's overview of [Your Apps](https://api.slack.com/apps) and create a new App.
2. Under **Add features and functionality**, select **Slash Commands** (or go to **Slash Commands** under **Features** in the sidebar).  
    Register the slash command(s) that you're building. You can use the same **Request URL** for all of them. The URL of the example below ends with `/slash`.
3. Under **Install your app to your workspace** click the button to **Install App to Workspace** (or go to **Install App** in the sidebar).
4. Go back to **Basic Information** and find the **Signing Secret** in the **App Credentials** section.  
    This is what you'll need for the `SLACK_SIGNING_SECRET` environment variable later.

## Usage

1. Import the package:  
    ```go
    import "htdvisser.dev/slash"
    ```
2. Create a new Router:  
    ```go
    r := slash.NewRouter(os.Getenv("SLACK_SIGNING_SECRET"))
    ```
3. Register a slash commands, for example `/ping`:  
    ```go
    r.RegisterCommand("/ping", func(ctx context.Context, req slash.Request) interface{} {
    		// ...
    })
    ```
4. Register the router to your ServeMux:  
    ```go
    http.Handle("/slash", r)
    ```

See the [example on Godoc](https://godoc.org/htdvisser.dev/slash#example-package) for more.

## Next Steps

- Go build something cool and showcase it on [the wiki](https://github.com/htdvisser/slash/wiki).
