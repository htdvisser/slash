# Slash

A router for Slack's [slash commands](https://api.slack.com/slash-commands).

## Preparations

1. Go to Slack's overview of [Your Apps](https://api.slack.com/apps) and create a new App.
2. After creating the App, find the **Signing Secret** in the **App Credentials** section.  
    This is what you'll need for the `SLACK_SIGNING_SECRET` environment variable later.
3. If you already know the URL where you plan to deploy your commands, you can go to **Slash Commands** under **Features**.  
    For each command that you create, set the **Request URL** to this URL. The URL of the example below ends with `/slash`.

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
    r.Register("/ping", func(ctx context.Context, req slash.Request) interface{} {
    		// ...
    })
    ```
4. Register the router to your ServeMux:  
    ```go
    http.Handle("/slash", r)
    ```

See the [example on Godoc](https://godoc.org/htdvisser.dev/slash#example-package) for more.
