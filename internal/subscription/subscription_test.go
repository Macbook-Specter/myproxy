package subscription

import (
	"testing"

	"myproxy.com/p/internal/config"
)

func TestSSParser(t *testing.T) {
	// 测试SS协议解析
	testCases := []struct {
		name     string
		input    string
		expected *config.Server
		wantErr  bool
	}{
		{
			name:  "basic SS format",
			input: "ss://YWVzLTI1Ni1nY206dGVzdHBhc3N3b3Jk@example.com:8388",
			expected: &config.Server{
				// ID 是通过 GenerateServerID 生成的 MD5 哈希值，我们不检查具体值
				Name:         "example.com:8388",
				Addr:         "example.com",
				Port:         8388,
				Username:     "testpassword",
				Password:     "testpassword",
				ProtocolType: "ss",
				SSMethod:     "aes-256-gcm",
			},
			wantErr: false,
		},
	}

	parser := &SSParser{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server, err := parser.Parse(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("SSParser.Parse() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr {
				// ID 是动态生成的，只要不为空即可
				if server.ID == "" {
					t.Errorf("Server.ID = '', want non-empty string")
				}
				if server.Name != tc.expected.Name {
					t.Errorf("Server.Name = %v, want %v", server.Name, tc.expected.Name)
				}
				if server.Addr != tc.expected.Addr {
					t.Errorf("Server.Addr = %v, want %v", server.Addr, tc.expected.Addr)
				}
				if server.Port != tc.expected.Port {
					t.Errorf("Server.Port = %v, want %v", server.Port, tc.expected.Port)
				}
				if server.ProtocolType != tc.expected.ProtocolType {
					t.Errorf("Server.ProtocolType = %v, want %v", server.ProtocolType, tc.expected.ProtocolType)
				}
				if server.SSMethod != tc.expected.SSMethod {
					t.Errorf("Server.SSMethod = %v, want %v", server.SSMethod, tc.expected.SSMethod)
				}
			}
		})
	}
}

func TestTrojanParser(t *testing.T) {
	// 测试Trojan协议解析
	testCases := []struct {
		name     string
		input    string
		expected *config.Server
		wantErr  bool
	}{
		{
			name:  "basic Trojan format",
			input: "trojan://testpassword@example.com:443#TestServer",
			expected: &config.Server{
				// ID 是通过 GenerateServerID 生成的 MD5 哈希值，我们不检查具体值
				Name:           "TestServer",
				Addr:           "example.com",
				Port:           443,
				Username:       "testpassword",
				Password:       "testpassword",
				ProtocolType:   "trojan",
				TrojanPassword: "testpassword",
			},
			wantErr: false,
		},
	}

	parser := &TrojanParser{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server, err := parser.Parse(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("TrojanParser.Parse() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr {
				// ID 是动态生成的，只要不为空即可
				if server.ID == "" {
					t.Errorf("Server.ID = '', want non-empty string")
				}
				if server.Name != tc.expected.Name {
					t.Errorf("Server.Name = %v, want %v", server.Name, tc.expected.Name)
				}
				if server.Addr != tc.expected.Addr {
					t.Errorf("Server.Addr = %v, want %v", server.Addr, tc.expected.Addr)
				}
				if server.Port != tc.expected.Port {
					t.Errorf("Server.Port = %v, want %v", server.Port, tc.expected.Port)
				}
				if server.ProtocolType != tc.expected.ProtocolType {
					t.Errorf("Server.ProtocolType = %v, want %v", server.ProtocolType, tc.expected.ProtocolType)
				}
				if server.TrojanPassword != tc.expected.TrojanPassword {
					t.Errorf("Server.TrojanPassword = %v, want %v", server.TrojanPassword, tc.expected.TrojanPassword)
				}
			}
		})
	}
}
