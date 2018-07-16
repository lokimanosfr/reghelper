package main

import (
	"fmt"
	"reflect"
	"testing"

	"golang.org/x/sys/windows/registry"
)

func Test_setParams(t *testing.T) {
	type args struct {
		fullPath string
		value    string
		param    []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "DWORD",
			args:    args{"HKEY_CURRENT_CONFIG\\Software\\Fonts", "DWORD_TEST", []string{"10"}},
			wantErr: false,
		},
		{
			name:    "QWORD",
			args:    args{"HKEY_CURRENT_CONFIG\\Software\\Fonts", "QWORD_TEST", []string{"11"}},
			wantErr: false,
		},
		{
			name:    "EXPAND_SZ",
			args:    args{"HKEY_CURRENT_CONFIG\\Software\\Fonts", "EXPAND_SZ_TEST", []string{"12"}},
			wantErr: false,
		},
		{
			name:    "MULTI_SZ",
			args:    args{"HKEY_CURRENT_CONFIG\\Software\\Fonts", "MULTI_SZ_TEST", []string{"13", "14"}},
			wantErr: false,
		},
		{
			name:    "SZ",
			args:    args{"HKEY_CURRENT_CONFIG\\Software\\Fonts", "SZ_TEST", []string{"15"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setParams(tt.args.fullPath, tt.args.value, tt.args.param); (err != nil) != tt.wantErr {
				t.Errorf("setParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getParams(t *testing.T) {

	err := setParams("HKEY_CURRENT_CONFIG\\Software\\Fonts", "EXPAND_SZ_TEST", []string{"hello great -Xmx=10G"})
	if err != nil {
		fmt.Println("set err + " + err.Error())
	}
	err = setParams("HKEY_CURRENT_CONFIG\\Software\\Fonts", "SZ_TEST", []string{"hello great -Xmx=10G hello  -Xmx=10Ggreat"})
	if err != nil {
		fmt.Println("set err + " + err.Error())
	}
	type args struct {
		fullPath string
		value    string
	}
	tests := []struct {
		name  string
		args  args
		want  interface{}
		want1 uint32
	}{
		{
			name:  "DWORD",
			args:  args{"HKEY_CURRENT_CONFIG\\Software\\Fonts", "DWORD_TEST"},
			want:  uint64(10),
			want1: registry.DWORD,
		},
		{
			name:  "QWORD",
			args:  args{"HKEY_CURRENT_CONFIG\\Software\\Fonts", "QWORD_TEST"},
			want:  uint64(11),
			want1: registry.QWORD,
		},
		{
			name:  "EXPAND_SZ",
			args:  args{"HKEY_CURRENT_CONFIG\\Software\\Fonts", "EXPAND_SZ_TEST"},
			want:  "12",
			want1: registry.EXPAND_SZ,
		},
		{
			name:  "MULTI_SZ",
			args:  args{"HKEY_CURRENT_CONFIG\\Software\\Fonts", "MULTI_SZ_TEST"},
			want:  []string{"13", "14"},
			want1: registry.MULTI_SZ,
		},
		{
			name:  "SZ",
			args:  args{"HKEY_CURRENT_CONFIG\\Software\\Fonts", "SZ_TEST"},
			want:  "15",
			want1: registry.SZ,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getParams(tt.args.fullPath, tt.args.value)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getParams() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getParams() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_replaceParams(t *testing.T) {
	err := setParams("HKEY_CURRENT_CONFIG\\Software\\Fonts", "EXPAND_SZ_TEST", []string{"hello great -Xmx=10G"})
	if err != nil {
		fmt.Println("set err + " + err.Error())
	}
	err = setParams("HKEY_CURRENT_CONFIG\\Software\\Fonts", "SZ_TEST", []string{"hello great -Xmx=10G hello  -Xmx=10Ggreat"})
	if err != nil {
		fmt.Println("set err + " + err.Error())
	}

	type args struct {
		fullPath string
		value    string
		param    []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "EXPAND_SZ",
			args: args{"HKEY_CURRENT_CONFIG\\Software\\Fonts", "EXPAND_SZ_TEST", []string{"-Xmx=.*=>-Xmx=2G"}},
			want: "hello great -Xmx=2G",
		},
		{
			name: "SZ",
			args: args{"HKEY_CURRENT_CONFIG\\Software\\Fonts", "SZ_TEST", []string{"-Xmx=.*=>-Xmx=1G"}},
			want: "hello great -Xmx=1G",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := replaceParams(tt.args.fullPath, tt.args.value, tt.args.param); got != tt.want {
				t.Errorf("replaceParams() = %v, want %v", got, tt.want)
			}
		})
	}
}
