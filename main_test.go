package main

import (
	"fmt"
	"reflect"
	"testing"

	"golang.org/x/sys/windows/registry"
)

func Test_getParams(t *testing.T) {

	_, _, err := setParams("HKEY_CURRENT_CONFIG\\Software\\Fonts", "EXPAND_SZ_TEST", []string{"12"})
	_, _, err = setParams("HKEY_CURRENT_CONFIG\\Software\\Fonts", "SZ_TEST", []string{"15"})

	_, _, err = setParams("HKEY_CURRENT_CONFIG\\Software\\Fonts", "MULTI_SZ_TEST", []string{"[13;14]"})

	_, _, err = setParams("HKEY_CURRENT_CONFIG\\Software\\Fonts", "DWORD_TEST", []string{"10"})

	_, _, err = setParams("HKEY_CURRENT_CONFIG\\Software\\Fonts", "QWORD_TEST", []string{"11"})
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
			want:  uint32(10),
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

func Test_getSplitedParams(t *testing.T) {
	type args struct {
		argSet string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "array",
			args: args{"[1;2;3;4]"},
			want: []string{"1", "2", "3", "4"},
		},
		{
			name: "string",
			args: args{"[1;2;3;4"},
			want: []string{"[1;2;3;4"},
		},
		{
			name: "has empty string",
			args: args{"[1,2,3;332;1,2,3;[123];;;;;,;\\;;,;]"},
			want: []string{"1,2,3", "332", "1,2,3", "[123]", ",", "\\", ","},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSplitedParams(tt.args.argSet); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSplitedParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createKey(t *testing.T) {
	path := "HKEY_LOCAL_MACHINE\\SOFTWARE\\Google\\Chrome"
	key := "NEWKEY"
	fmt.Println(deleteKey(path, key))
	type args struct {
		fullPath string
		underKey string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Create key",
			args: args{path, key},
			want: path + "\\" + key + " created",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createKey(tt.args.fullPath, tt.args.underKey); got != tt.want {
				t.Errorf("createKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setParams(t *testing.T) {
	type args struct {
		fullPath string
		value    string
		param    []string
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		typ     uint32
		wantErr bool
	}{
		{
			name:    "DWORD",
			args:    args{"HKEY_CURRENT_CONFIG\\Software\\Fonts", "DWORD_TEST", []string{"10"}},
			wantErr: false,
			want:    uint32(10),
			typ:     4,
		},
		{
			name:    "QWORD",
			args:    args{"HKEY_CURRENT_CONFIG\\Software\\Fonts", "QWORD_TEST", []string{"11"}},
			wantErr: false,
			want:    uint64(11),
			typ:     11,
		},
		{
			name:    "EXPAND_SZ",
			args:    args{"HKEY_CURRENT_CONFIG\\Software\\Fonts", "EXPAND_SZ_TEST", []string{"12"}},
			wantErr: false,
			want:    string("12"),
			typ:     2,
		},
		{
			name:    "MULTI_SZ",
			args:    args{"HKEY_CURRENT_CONFIG\\Software\\Fonts", "MULTI_SZ_TEST", []string{"[13;14]"}},
			wantErr: false,
			want:    []string{"13", "14"},
			typ:     7,
		},
		{
			name:    "SZ",
			args:    args{"HKEY_CURRENT_CONFIG\\Software\\Fonts", "SZ_TEST", []string{"15"}},
			wantErr: false,
			want:    string("15"),
			typ:     1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := setParams(tt.args.fullPath, tt.args.value, tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("setParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				switch got1 {
				case 1:
					t.Errorf("setParams() got = %v, want %v", getStringFromInterface(got), tt.want)
				case 7:
					t.Errorf("setParams() got = %v, want %v", getStringsFromInterface(got), tt.want)
				case 2:
					t.Errorf("setParams() got = %v, want %v", getStringFromInterface(got), tt.want)
				case 4:
					t.Errorf("setParams() got = %v, want %v", getUint32FromInterface(got), tt.want)
				case 11:
					t.Errorf("setParams() got = %v, want %v", getUint64FromInterface(got), tt.want)
				}
			}
			if got1 != tt.typ {
				t.Errorf("setParams() got1 = %v, want %v", got1, tt.typ)
			}
		})
	}
}

// func Test_replaceParams(t *testing.T) {
// 	_, _, err := setParams("HKEY_CURRENT_CONFIG\\Software\\Fonts", "EXPAND_SZ_TEST", []string{"hello great -Xmx=10G"})
// 	_, _, err = setParams("HKEY_CURRENT_CONFIG\\Software\\Fonts", "SZ_TEST", []string{"hello great -Xmx=10G hello  -Xmx=10Ggreat"})
// 	_, _, err = setParams("HKEY_CURRENT_CONFIG\\Software\\Fonts", "MULTI_SZ_TEST", []string{"[hello great ;;;-Xmx=10G hello  -Xmx=10Ggreat;hello great -Xmx=10G hello  -Xmx=10Ggreat;hello great -Xmx=10G hello  -Xmx=10Ggreat]"})
// 	type args struct {
// 		fullPath string
// 		value    string
// 		param    []string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantWas interface{}
// 		wantNow interface{}
// 		wantTyp uint32
// 		wantErr bool
// 	}{
// 		{
// 			name: "EXPAND_SZ",
// 			args: args{"HKEY_CURRENT_CONFIG\\Software\\Fonts", "EXPAND_SZ_TEST", []string{"-Xmx=.*=>-Xmx=2G"}},
// 			want: "hello great -Xmx=2G replace successful",
// 		},
// 		{
// 			name: "SZ",
// 			args: args{"HKEY_CURRENT_CONFIG\\Software\\Fonts", "SZ_TEST", []string{"-Xmx=.*=>-Xmx=1G"}},
// 			want: "hello great -Xmx=1G replace successful",
// 		},
// 		{
// 			name: "MULTI_SZ",
// 			args: args{"HKEY_CURRENT_CONFIG\\Software\\Fonts", "MULTI_SZ_TEST", []string{"great=>OLO"}},
// 			want: "replace successful",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			gotWas, gotNow, gotTyp, err := replaceParams(tt.args.fullPath, tt.args.value, tt.args.param)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("replaceParams() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(gotWas, tt.wantWas) {
// 				t.Errorf("replaceParams() gotWas = %v, want %v", gotWas, tt.wantWas)
// 			}
// 			if !reflect.DeepEqual(gotNow, tt.wantNow) {
// 				t.Errorf("replaceParams() gotNow = %v, want %v", gotNow, tt.wantNow)
// 			}
// 			if gotTyp != tt.wantTyp {
// 				t.Errorf("replaceParams() gotTyp = %v, want %v", gotTyp, tt.wantTyp)
// 			}
// 		})
// 	}
// }
