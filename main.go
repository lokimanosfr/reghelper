package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/sys/windows/registry"
)



type CommandArgs struct {
	path        string
	value       string
	set         string
	replace     string
	get         bool
	contain     string
	help        bool
	valType     string
	deleteKey   string
	deleteValue string
	createKey   string
}

var (
	args     = CommandArgs{}
	regTypes = map[string]uint32{
		"TEST":		registry.
		"DWORD":     registry.DWORD,
		"QWORD":     registry.QWORD,
		"SZ":        registry.SZ,
		"MULTI_SZ":  registry.MULTI_SZ,
		"EXPAND_SZ": registry.EXPAND_SZ,
	}
	//Access is a map for access
	Access = map[string]uint32{
		"read":   registry.QUERY_VALUE,
		"write":  registry.QUERY_VALUE | registry.SET_VALUE,
		"create": registry.QUERY_VALUE | registry.CREATE_SUB_KEY,
		"delete": registry.QUERY_VALUE | registry.SET_VALUE,
	}
)

func init() {

	flag.StringVar(&args.path, "path", "", "Set the path to registry key")
	flag.StringVar(&args.value, "value", "", "Set which value use")
	flag.StringVar(&args.deleteKey, "delkey", "", "Will delete subkey(value) from path")
	flag.StringVar(&args.deleteValue, "delval", "", "Will delete value of subkey ")
	flag.StringVar(&args.set, "set", "", "Set params to value")
	flag.StringVar(&args.replace, "replace", "", "Replace param or substring to another")
	flag.BoolVar(&args.get, "get", false, "Get parametrs of value")
	flag.StringVar(&args.contain, "contain", "", "Chek that the value contain param")
	flag.StringVar(&args.valType, "type", "", "Used when need create value")
	flag.StringVar(&args.createKey, "createkey", "", "create subkey")
	flag.BoolVar(&args.help, "help", false, "Show usage")
	flag.Parse()
	if args.help {
		fmt.Println("Usage:")
		fmt.Println("-path <Path to key> -value <Value Name> -set <param>\t| Set params to value")
		fmt.Println("-path <Path to key> -value <Value Name> -type <value type> -set <param>\t| Set params to value even it doesn't exist ( will create value)")
		fmt.Println("-path <Path to key> -value <Value Name> -set \"<[param;param]>\"\t|Usage for REG_MULTI_SZ. Set params to value")
		fmt.Println("-path <Path to key> -value <Value Name> type <value type> -set <[param;param]>\t| Usage for REG_MULTI_SZ. Set params to value even it doesn't exist ( will create value)")
		fmt.Println("-path <Path to key> -value <Value Name> -replace <param>=><param>\t| Usage for replace param or substring to another. ")
		fmt.Println("-path <Path to key> -value <Value Name> -get\t| Get parametrs of value")
		fmt.Println("-path <Path to key> -value <Value Name> -contain <param>\t| fing param and return true if it's finded or false if not")
		fmt.Println("-path <Path to key>  -delkey <key>\t| Will delete key from path")
		fmt.Println("-path <Path to key>  -delval <value>\t| Will delete value from path ")
		fmt.Println("-path <Path to key>  -createkey <key>\t| Will create key ")
		fmt.Println("-value <DWORD,QWORD,SZ,MULTI_SZ,EXPAND_SZ>\t| Types for params ")
		os.Exit(0)
	}

}

//main get fdfdfd
func main() {

	if args.path == "" {
		fmt.Println("invalid path and value, use -help")
		return
	}
	switch arg := args.chekArgs(); arg {
	case "set":

		i, typ, err := setParams(args.path, args.value, []string{args.set})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		switch typ {
		case 11:
			strUint := strconv.FormatUint(getUint64FromInterface(i), 10)
			fmt.Println("Set " + strUint + " to " + args.path + "\\" + args.value + " successful")
		case 4:
			strUint := strconv.FormatUint(uint64(getUint32FromInterface(i)), 10)
			fmt.Println("Set " + strUint + " to " + args.path + "\\" + args.value + " successful")
		case 1, 2:
			fmt.Println("Set " + getStringFromInterface(i) + " to " + args.path + "\\" + args.value + " successful")
		case 7:
			arr := strings.Join(getStringsFromInterface(i), ";")
			fmt.Println("Set " + arr + " to " + args.path + "\\" + args.value + " successful")
		default:

		}

	case "replace":
		was, now, typ, err := replaceParams(args.path, args.value, []string{args.replace})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println("Replaced successful!")
		switch typ {
		case regTypes["SZ"], regTypes["EXPAND_SZ"]:
			fmt.Println("Was: " + getStringFromInterface(was))
			fmt.Println("Now: " + getStringFromInterface(now))
		case regTypes["MULTI_SZ"]:
			fmt.Println("Was: " + strings.Join(getStringsFromInterface(was), ";"))
			fmt.Println("Now: " + strings.Join(getStringsFromInterface(now), ";"))

		}
	case "get":
		i, typ := getParams(args.path, args.value)
		switch typ {
		case 11:
			fmt.Println(getUint64FromInterface(i))
		case 4:
			fmt.Println(getUint32FromInterface(i))
		case 1, 2:
			fmt.Println(getStringFromInterface(i))
		case 7:
			fmt.Println(getStringsFromInterface(i))
		default:
			fmt.Println(args.path + "\\" + args.value + " is not exist")

		}
	case "contain":

	case "create":
		fmt.Println(createKey(args.path, args.createKey))

	case "delkey":
		fmt.Println(deleteKey(args.path, args.deleteKey))
	case "delval":
		fmt.Println(deleteValue(args.path, args.deleteValue))
	default:
		fmt.Println("invaid argument, use -help")
	}

}

func getSplitedParams(argSet string) []string {
	var spl []string
	var out []string
	if string(argSet[0]) == "[" && string(argSet[len(argSet)-1]) == "]" && strings.Contains(argSet, ";") {
		spl = strings.Split(string(argSet[1:len(argSet)-1]), ";")
		for _, p := range spl {
			if p != "" {
				out = append(out, p)
			}
		}
	} else {
		out = append(out, argSet)
	}
	return out
}

func (args *CommandArgs) chekArgs() string {
	if args.set != "" {
		return "set"
	}
	if args.replace != "" {
		return "replace"
	}
	if args.get != false {
		return "get"
	}
	if args.contain != "" {
		return "contain"
	}
	if args.deleteKey != "" {
		return "delkey"
	}
	if args.deleteValue != "" {
		return "delval"
	}
	if args.createKey != "" {
		return "create"
	}

	return ""
}

func getExistValueType(fullPath, value string) uint32 {

	key := openKey(fullPath, Access["read"])
	_, typ, err := key.GetValue(value, make([]byte, 0, 0))
	if err != nil {
		return 0
	}
	defer key.Close()
	return typ

}

func getStringFromInterface(i interface{}) string {
	return i.(string)
}
func getUint64FromInterface(i interface{}) uint64 {
	v := reflect.ValueOf(i)
	return v.Interface().(uint64)
}
func getUint32FromInterface(i interface{}) uint32 {
	v := reflect.ValueOf(i)
	return v.Interface().(uint32)
}
func getStringsFromInterface(i interface{}) []string {
	src := reflect.ValueOf(i)
	srcArr := src.Interface().([]string)
	return srcArr

}

func replaceParams(fullPath, value string, param []string) (was interface{}, now interface{}, typ uint32, err error) {
	argsToReplace := strings.Split(param[0], "=>")
	typ = getExistValueType(fullPath, value)
	if typ == 0 {
		return nil, nil, 0, errors.New(value + " is not exist")
	}

	srcInterface, _ := getParams(fullPath, value)
	switch typ {
	case 1:
		src := getStringFromInterface(srcInterface)
		re := regexp.MustCompile(argsToReplace[0])
		ss := re.ReplaceAllString(src, argsToReplace[1])
		i, typ, err := setParams(fullPath, value, []string{ss})
		return srcInterface, i, typ, err
	case 2:
		src := getStringFromInterface(srcInterface)
		re := regexp.MustCompile(argsToReplace[0])
		ss := re.ReplaceAllString(src, argsToReplace[1])
		i, typ, err := setParams(fullPath, value, []string{ss})
		return srcInterface, i, typ, err
	case 7:
		src := getStringsFromInterface(srcInterface)
		re := regexp.MustCompile(argsToReplace[0])
		var outArr []string
		for _, line := range src {
			ss := re.ReplaceAllString(line, argsToReplace[1])
			outArr = append(outArr, ss)
		}
		i, typ, err := setParams(fullPath, value, []string{"[" + strings.Join(outArr, ";") + "]"})
		return srcInterface, i, typ, err
	default:
		return nil, nil, 0, errors.New("Can't replace DWORD,QWORD,BINARY parametrs of value, only set")
	}

}

func set(fullPath, value string, valType uint32, param []string) (interface{}, error) {
	var err error
	key := openKey(fullPath, Access["write"])
	switch valType {
	case 1:
		err = key.SetStringValue(value, param[0])
	case 2:
		err = key.SetExpandStringValue(value, param[0])
	case 4:
		i, err2 := strconv.ParseUint(param[0], 10, 32)
		if err2 != nil {
			return nil, err2
		}
		err = key.SetDWordValue(value, uint32(i))
	case 11:
		i, err2 := strconv.ParseUint(param[0], 10, 64)
		if err2 != nil {
			return nil, err2
		}
		err = key.SetQWordValue(value, uint64(i))
	case 7:
		err = key.SetStringsValue(value, param)

	}
	key.Close()
	i, _ := getParams(fullPath, value)
	return i, err
}

func setParams(fullPath, value string, param []string) (interface{}, uint32, error) {
	spl := getSplitedParams(param[0])
	typ := getExistValueType(fullPath, value)
	var i interface{}
	var err error
	if typ == 0 {
		if args.valType != "" {
			i, err = set(fullPath, value, regTypes[args.valType], spl)
			return i, typ, err
		}
		fmt.Println("Value is not exist, use -help to know how create value")

	} else {
		i, err = set(fullPath, value, typ, spl)
	}
	return i, typ, err
}

func deleteKey(fullPath, path string) string {
	key := openKey(fullPath, Access["delete"])
	err := registry.DeleteKey(key, path)
	if err != nil {
		return "Key " + path + " not deleted:\n" + err.Error()
	}
	return "Key " + path + " delete successful"
}

func deleteValue(fullPath, value string) string {
	key := openKey(fullPath, Access["delete"])
	err := key.DeleteValue(value)
	if err != nil {
		return "Value " + value + " not deleted:\n" + err.Error()
	}
	return "Value " + value + " deleted successful"

}

func createKey(fullPath, subKey string) string {
	key := openKey(fullPath, Access["create"])

	_, opened, err := registry.CreateKey(key, subKey, registry.CREATE_SUB_KEY)
	if err != nil {
		return err.Error()
	}
	if opened {
		return "Key alredy exist"
	}
	return fullPath + "\\" + subKey + " created"
}

func getParams(fullPath, value string) (interface{}, uint32) {
	typ := getExistValueType(fullPath, value)
	if typ == 0 {
		return nil, 0
	}
	key := openKey(fullPath, Access["read"])
	switch typ {
	case 1, 2:
		str, _, _ := key.GetStringValue(value)
		return str, typ
	case 4:
		num, _, _ := key.GetIntegerValue(value)
		return uint32(num), typ
	case 11:
		num, _, _ := key.GetIntegerValue(value)
		return num, typ
	case 7:
		strArr, _, _ := key.GetStringsValue(value)
		return strArr, typ
	}
	key.Close()
	return nil, 0
}

func getHKEY(fullPath string) string {
	return strings.Split(fullPath, "\\")[0]
}
func getKeyPath(fullPath string) string {
	return strings.Join(strings.Split(fullPath, "\\")[1:], "\\")
}

func openKey(fullPath string, access uint32) registry.Key {
	var k registry.Key
	var err error

	switch hkey := getHKEY(fullPath); hkey {
	case "HKEY_LOCAL_MACHINE":
		k, err = registry.OpenKey(registry.LOCAL_MACHINE, getKeyPath(fullPath), access)
	case "HKEY_CURRENT_USER":
		k, err = registry.OpenKey(registry.CURRENT_USER, getKeyPath(fullPath), access)
	case "HKEY_CLASSES_ROOT":
		k, err = registry.OpenKey(registry.CLASSES_ROOT, getKeyPath(fullPath), access)
	case "HKEY_USERS":
		k, err = registry.OpenKey(registry.USERS, getKeyPath(fullPath), access)
	case "HKEY_CURRENT_CONFIG":
		k, err = registry.OpenKey(registry.CURRENT_CONFIG, getKeyPath(fullPath), access)
	}
	if err != nil {
		log.Fatal(err.Error() + "\npath = " + fullPath)
	}

	return k
}
