# Slash

A router for Slack's [slash commands](https://api.slack.com/slash-commands).

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
