# 🌀 Kron

**Kron** is an open-source automation framework written in Go.  
It aims to be the **Swiss army knife for DevOps and SREs** — a single CLI that can automate everything from editing configs to setting up observability stacks or managing GitOps pipelines.

Right now, Kron provides the **foundation** — a modular CLI where anyone can contribute new automation commands.

---

## 🎯 Vision

Kron's long-term goal is to be your all-in-one DevOps assistant:
- Manage YAML/JSON across environments
- Automate GitHub/GitLab operations (PRs, branches, merges)
- Spin up monitoring stacks (Prometheus, Grafana) locally or remotely
- Manage cloud infra (AWS, GCP, OCI)
- Offer plugin-style command extensions

Kron = **Declarative. Extensible. Fast. Written in Go.**

---

## 🧱 Current Status

Kron is in **early development**.  
It currently provides:
- Basic CLI structure built with [Cobra](https://github.com/spf13/cobra)
- Example command (`yaml`) for testing new command modules
- Open design ready for contributors to add subcommands

---

## 🧑‍💻 How to Contribute a Command

We’re building Kron *together*.  
If you’ve ever thought “this could be automated,” Kron is your sandbox.

1. Fork the repo:
   ```bash
   git clone https://github.com/<your-username>/kron.git
   cd kron
   ```

2. Generate a new command:
   ```bash
   go install github.com/spf13/cobra-cli@latest
   cobra-cli add <command-name>
   ```

3. Implement your logic in `cmd/<command-name>.go`  
   Example:
   ```go
   var myCmd = &cobra.Command{
     Use:   "hello",
     Short: "Prints Hello World",
     Run: func(cmd *cobra.Command, args []string) {
       fmt.Println("Hello from Kron!")
     },
   }
   ```

4. Register your command in `init()` and test:
   ```bash
   go run main.go <command-name>
   ```

---

## 🛠 Installation

### From Source
```bash
git clone https://github.com/kroniar/kron.git
cd kron
go run main.go
```

---

## 🤝 Contributing

You can:
- Add a new command (`cmd/` directory)
- Improve CLI help text and UX
- Write tests or docs
- Propose architectural improvements

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed instructions.

---

## 🧭 Roadmap

- [x] Project bootstrapped with Cobra
- [ ] CLI plugin loader
- [ ] Built-in YAML/JSON operations
- [ ] GitOps commands
- [ ] Observability setup (Prometheus, Grafana)
- [ ] Multi-cloud integrations

---

## 📄 License

Licensed under the [MIT License](LICENSE).

---

## 💡 About

Maintained by [Neeraj Gupta](https://github.com/6620913) under the [Kroniar](https://github.com/kroniar) organization.  
Contributions welcome — **make automation fun again.**
