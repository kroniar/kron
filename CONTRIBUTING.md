# 🧩 Contributing to Kron

Kron is designed as an **open automation platform**.  
At this stage, the best way to contribute is to **add new CLI commands** — each command can automate something DevOps or SRE-related.

---

## 🧑‍💻 Local Setup

1. Fork and clone the repo:
   ```bash
   git clone https://github.com/<your-username>/kron.git
   cd kron
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Run the app:
   ```bash
   go run main.go
   ```

---

## 🧱 Add a New Command

We use [Cobra](https://github.com/spf13/cobra) for CLI structure.

To add a new command:
```bash
go install github.com/spf13/cobra-cli@latest
cobra-cli add mycommand
```

Then open `cmd/mycommand.go` and edit:
```go
var mycommandCmd = &cobra.Command{
  Use:   "mycommand",
  Short: "What your command does",
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("My command is running!")
  },
}

func init() {
  rootCmd.AddCommand(mycommandCmd)
}
```

---

## 🧠 Example Ideas for Commands

- `yaml` → modify multiple YAMLs at once  
- `git` → automate PR creation, tagging, or merging  
- `prometheus` → deploy local Prometheus stack  
- `grafana` → configure dashboards automatically  
- `cloud` → manage AWS/GCP resources from CLI  
- `infra` → provision local environments with Docker  

Each of these could live as independent commands under `cmd/`.

---

## 🧼 Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Run `go fmt ./...` before committing
- Keep commits small and meaningful

---

## 🧪 Testing

We’ll eventually add full Go tests, but for now, manual testing via:
```bash
go run main.go <command>
```

---

## 🫶 Community

Be kind, collaborative, and constructive — Kron is open to all contributors.  
If you add something useful, it’ll be merged and credited.

You can discuss ideas by opening **Issues** or **PRs**.

---

Let’s automate the world — one CLI command at a time. 🌍
