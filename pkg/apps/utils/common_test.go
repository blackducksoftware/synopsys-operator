package utils

import "testing"

func TestGetResourceName(t *testing.T) {
	type args struct {
		name        string
		appName     string
		defaultName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no appName",
			args: args{
				name:        "name",
				appName:     "",
				defaultName: "defaultName",
			},
			want: "name-defaultName",
		},
		{
			name: "appName, no defaultName",
			args: args{
				name:        "name",
				appName:     "appName",
				defaultName: "",
			},
			want: "name-appName",
		},
		{
			name: "appName, defaultName",
			args: args{
				name:        "name",
				appName:     "appName",
				defaultName: "defaultName",
			},
			want: "name-appName-defaultName",
		},
		// now not covered
		{
			name: "no appName, no defaultName",
			args: args{
				name:        "name",
				appName:     "",
				defaultName: "",
			},
			want: "name-",
		},
		{
			name: "all empty",
			args: args{
				name:        "",
				appName:     "",
				defaultName: "",
			},
			want: "-",
		},
		{
			name: "just defaultName",
			args: args{
				name:        "",
				appName:     "",
				defaultName: "defaultName",
			},
			want: "-defaultName",
		},
		{
			name: "just appName",
			args: args{
				name:        "",
				appName:     "appName",
				defaultName: "",
			},
			want: "-appName",
		},
		{
			name: "no name",
			args: args{
				name:        "",
				appName:     "appName",
				defaultName: "defaultName",
			},
			want: "-appName-defaultName",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetResourceName(tt.args.name, tt.args.appName, tt.args.defaultName); got != tt.want {
				t.Errorf("GetResourceName() = %v, want %v", got, tt.want)
			}
		})
	}
}
