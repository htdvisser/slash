workflow "Test and Build" {
  on = "push"
  resolves = ["Test"]
}

action "Test" {
  uses = "docker://golang:1.12-stretch"
  args = ["go", "test", "-cover", "./..."]
}
