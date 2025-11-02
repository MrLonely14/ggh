package tunnel

import (
	"testing"
)

func TestTunnelValidation(t *testing.T) {
	tests := []struct {
		name    string
		tunnel  Tunnel
		wantErr bool
	}{
		{
			name: "Valid local tunnel",
			tunnel: Tunnel{
				Name:       "test-local",
				Type:       TypeLocal,
				LocalPort:  8080,
				RemoteHost: "localhost",
				RemotePort: 80,
			},
			wantErr: false,
		},
		{
			name: "Valid remote tunnel",
			tunnel: Tunnel{
				Name:       "test-remote",
				Type:       TypeRemote,
				LocalPort:  8080,
				RemoteHost: "localhost",
				RemotePort: 3000,
			},
			wantErr: false,
		},
		{
			name: "Valid dynamic tunnel",
			tunnel: Tunnel{
				Name:      "test-dynamic",
				Type:      TypeDynamic,
				LocalPort: 1080,
			},
			wantErr: false,
		},
		{
			name: "Invalid - empty name",
			tunnel: Tunnel{
				Type:       TypeLocal,
				LocalPort:  8080,
				RemoteHost: "localhost",
				RemotePort: 80,
			},
			wantErr: true,
		},
		{
			name: "Invalid - bad port",
			tunnel: Tunnel{
				Name:       "test",
				Type:       TypeLocal,
				LocalPort:  99999,
				RemoteHost: "localhost",
				RemotePort: 80,
			},
			wantErr: true,
		},
		{
			name: "Invalid - local missing remote host",
			tunnel: Tunnel{
				Name:      "test",
				Type:      TypeLocal,
				LocalPort: 8080,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tunnel.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestToSSHArgs(t *testing.T) {
	tests := []struct {
		name     string
		tunnel   Tunnel
		wantArgs []string
		wantErr  bool
	}{
		{
			name: "Local forwarding",
			tunnel: Tunnel{
				Name:       "test",
				Type:       TypeLocal,
				LocalPort:  8080,
				RemoteHost: "localhost",
				RemotePort: 80,
			},
			wantArgs: []string{"-L", "8080:localhost:80"},
			wantErr:  false,
		},
		{
			name: "Remote forwarding",
			tunnel: Tunnel{
				Name:       "test",
				Type:       TypeRemote,
				LocalPort:  8080,
				RemoteHost: "localhost",
				RemotePort: 3000,
			},
			wantArgs: []string{"-R", "8080:localhost:3000"},
			wantErr:  false,
		},
		{
			name: "Dynamic forwarding",
			tunnel: Tunnel{
				Name:      "test",
				Type:      TypeDynamic,
				LocalPort: 1080,
			},
			wantArgs: []string{"-D", "1080"},
			wantErr:  false,
		},
		{
			name: "Local with bind address",
			tunnel: Tunnel{
				Name:        "test",
				Type:        TypeLocal,
				LocalPort:   8080,
				RemoteHost:  "localhost",
				RemotePort:  80,
				BindAddress: "127.0.0.1",
			},
			wantArgs: []string{"-L", "127.0.0.1:8080:localhost:80"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, err := tt.tunnel.ToSSHArgs()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToSSHArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(args) != len(tt.wantArgs) {
					t.Errorf("ToSSHArgs() = %v, want %v", args, tt.wantArgs)
					return
				}
				for i := range args {
					if args[i] != tt.wantArgs[i] {
						t.Errorf("ToSSHArgs()[%d] = %v, want %v", i, args[i], tt.wantArgs[i])
					}
				}
			}
		})
	}
}

func TestParseSSHFlag(t *testing.T) {
	tests := []struct {
		name    string
		flag    string
		want    *Tunnel
		wantErr bool
	}{
		{
			name: "Local forwarding",
			flag: "-L 8080:localhost:80",
			want: &Tunnel{
				Type:       TypeLocal,
				LocalPort:  8080,
				RemoteHost: "localhost",
				RemotePort: 80,
			},
			wantErr: false,
		},
		{
			name: "Remote forwarding",
			flag: "-R 8080:localhost:3000",
			want: &Tunnel{
				Type:       TypeRemote,
				LocalPort:  8080,
				RemoteHost: "localhost",
				RemotePort: 3000,
			},
			wantErr: false,
		},
		{
			name: "Dynamic forwarding",
			flag: "-D 1080",
			want: &Tunnel{
				Type:      TypeDynamic,
				LocalPort: 1080,
			},
			wantErr: false,
		},
		{
			name: "Local with bind address",
			flag: "-L 127.0.0.1:8080:localhost:80",
			want: &Tunnel{
				Type:        TypeLocal,
				LocalPort:   8080,
				RemoteHost:  "localhost",
				RemotePort:  80,
				BindAddress: "127.0.0.1",
			},
			wantErr: false,
		},
		{
			name:    "Invalid flag",
			flag:    "-X invalid",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSSHFlag(tt.flag)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSSHFlag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Type != tt.want.Type {
					t.Errorf("ParseSSHFlag() Type = %v, want %v", got.Type, tt.want.Type)
				}
				if got.LocalPort != tt.want.LocalPort {
					t.Errorf("ParseSSHFlag() LocalPort = %v, want %v", got.LocalPort, tt.want.LocalPort)
				}
				if got.RemoteHost != tt.want.RemoteHost {
					t.Errorf("ParseSSHFlag() RemoteHost = %v, want %v", got.RemoteHost, tt.want.RemoteHost)
				}
				if got.RemotePort != tt.want.RemotePort {
					t.Errorf("ParseSSHFlag() RemotePort = %v, want %v", got.RemotePort, tt.want.RemotePort)
				}
				if got.BindAddress != tt.want.BindAddress {
					t.Errorf("ParseSSHFlag() BindAddress = %v, want %v", got.BindAddress, tt.want.BindAddress)
				}
			}
		})
	}
}

func TestTunnelsToSSHArgs(t *testing.T) {
	tunnels := []Tunnel{
		{
			Name:       "local-tunnel",
			Type:       TypeLocal,
			LocalPort:  8080,
			RemoteHost: "localhost",
			RemotePort: 80,
		},
		{
			Name:      "dynamic-tunnel",
			Type:      TypeDynamic,
			LocalPort: 1080,
		},
	}

	args, err := TunnelsToSSHArgs(tunnels)
	if err != nil {
		t.Errorf("TunnelsToSSHArgs() error = %v", err)
		return
	}

	expected := []string{"-L", "8080:localhost:80", "-D", "1080"}
	if len(args) != len(expected) {
		t.Errorf("TunnelsToSSHArgs() = %v, want %v", args, expected)
		return
	}

	for i := range args {
		if args[i] != expected[i] {
			t.Errorf("TunnelsToSSHArgs()[%d] = %v, want %v", i, args[i], expected[i])
		}
	}
}
