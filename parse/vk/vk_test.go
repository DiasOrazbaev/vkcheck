package vk

import "testing"

func TestSendLog(t *testing.T) {
	type args struct {
		username string
		password string
	}
	tests := []struct {
		name string
		args args
	}{
		{"Simple", args{"someusername", "somepassword"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SendLog(tt.args.username, tt.args.password)
		})
	}
}
