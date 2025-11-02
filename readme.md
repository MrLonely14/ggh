<p align="center">
    <img width="80" height="70" src="./assets/ggh.png" alt="GGH logo">
</p>
<h1 align="center"/>GGH</h1>

<p align="center"><i>Recall your SSH sessions</i></p>

<p align="center"><img src="./assets/ggh.gif" alt="GGH Demo"></p>


## Install

Run one of the following script, or download the latest binary from the [releases](https://github.com/BlackOrder/ggh/releases) page.

```shell
# Unix based
curl https://raw.githubusercontent.com/BlackOrder/ggh/blackorder/install/unix.sh | sh

# Windows 
powershell -c "irm https://raw.githubusercontent.com/BlackOrder/ggh/blackorder/install/windows.ps1 | iex"

# Go
go install github.com/BlackOrder/ggh@blackorder
```

## Usages

### Basic SSH Connection

```shell
# Use it just like you're using SSH
ggh root@server.com
ggh root@server.com -p2440

# Run it with no arguments to get interactive list of the previous sessions
ggh

# Run it with - to get interactive list of all of your ~/.ssh/config listing
ggh -

# Run it with - STRING to get interactive filtered list of your ~/.ssh/config listing
ggh - stage
ggh - meta-servers

# To get non-interactive list of history and config, run
ggh --config
ggh --history
```

### Port Forwarding Tunnels

GGH includes comprehensive tunnel management for SSH port forwarding:

```shell
# Manage tunnels interactively (create, edit, delete, select)
ggh tunnels

# List all saved tunnels
ggh --tunnels

# Select tunnels and apply to SSH connection
ggh -t
```

#### Tunnel Types

GGH supports all three SSH port forwarding types:

- **Local Forwarding (-L)**: Forward local port to remote destination
  - Example: `8080:localhost:80` - Access remote port 80 via local port 8080

- **Remote Forwarding (-R)**: Forward remote port to local destination
  - Example: `8080:localhost:3000` - Expose local port 3000 on remote port 8080

- **Dynamic Forwarding (-D)**: SOCKS proxy for dynamic port forwarding
  - Example: `1080` - Create SOCKS proxy on port 1080

#### Interactive Tunnel Management

When you run `ggh tunnels`, you can:

- **Create new tunnels** (`n` key): Define reusable tunnel configurations
- **Edit tunnels** (`e` key): Modify existing tunnel settings
- **Delete tunnels** (`d` key): Remove tunnels you no longer need
- **Select tunnels** (Space/Enter): Choose tunnels to apply to connections
- **Filter tunnels** (`/` key): Search through your tunnel list

All tunnels are saved in `~/.ggh/tunnels.json` for easy reuse.

### GGH is NOT replacing SSH

In fact, GGH won't work if SSH is not installed or isn't available in your system's path.

GGH is meant to act as a lightweight, fast wrapper around your SSH commands.