package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/sys/windows/registry"
)

type commandArgs struct {
	path    string
	value   string
	set     string
	replace string
	get     string
	contain string
	help    bool
}

const (
	GET = "get"
	SET = "set"
)

var (
	args = commandArgs{}
)

func init() {

	flag.StringVar(&args.path, "path", "", "Set the path to registry key")
	flag.StringVar(&args.value, "value", "", "Set which value use")
	flag.StringVar(&args.set, "set", "", "Set params to value")
	flag.StringVar(&args.replace, "replace", "", "Repalce param or substring to another")
	flag.StringVar(&args.get, "get", "", "Get parametrs of value")
	flag.StringVar(&args.contain, "contain", "", "Chek that the value contain param")
	flag.BoolVar(&args.help, "help", false, "Show usage")
	flag.Parse()
	if args.help {
		fmt.Println("Usage:")
		fmt.Println("-path <Path to key> -val <Value Name> -set <param>\t	| Set params to value")
		fmt.Println("-path <Path to key> -val <Value Name> -set <[param;param]>\t| Usage for REG_MULTI_SZ. Set params to value")
		fmt.Println("-path <Path to key> -val <Value Name> -replace <param>=><param>\t| Usage for replace param or substring to another. ")
		fmt.Println("-path <Path to key> -val <Value Name> -get\t		| Get parametrs of value")
		fmt.Println("-path <Path to key> -val <Value Name> -contain <param>\t	| fing param and return true if it's finded or false if not")
		os.Exit(0)
	}

}

func main() {

	str := "Hello fucking world good world buy macbook world"
	fndstr := "w*.rl.?"
	re := regexp.MustCompile(fndstr)
	ss := re.ReplaceAllString(str, "BAD")
	fmt.Println(ss)

	return

	if args.path == "" || args.value == "" {
		fmt.Println("invalid path and value, use -help")
		return
	}
	switch arg := args.chekArgs(); arg {
	case "set":
		spl := getSplitedParams()
		err := setParams(args.path, args.value, spl)
		if err != nil {
			fmt.Println(err.Error())
		}
	case "replace":
		fmt.Println("USE SET")
	case "get":
		// i, t := getParams(args.path, args.value)
	case "contain":
		fmt.Println("USE SET")
	default:
		fmt.Println("invaid argument, use -help")
	}

}

func getSplitedParams() []string {
	var spl []string
	if string(args.set[0]) == "[" && string(args.set[len(args.set)-1]) == "]" {
		spl := strings.Split(string(args.set[1:len(args.set)-1]), ";")
		setParams(args.path, args.value, spl)
	} else {
		spl = append(spl, args.set)

	}
	return spl
}

func (args *commandArgs) chekArgs() string {
	if args.set != "" {
		return "set"
	}
	if args.replace != "" {
		return "replace"
	}
	if args.get != "" {
		return "get"
	}
	if args.contain != "" {
		return "contain"
	}
	return ""
}

func getValueType(fullPath, value string) uint32 {

	key := openKey(fullPath, GET)
	_, typ, err := key.GetValue(value, make([]byte, 0, 0))
	if err != nil {
		fmt.Println(err.Error())
		return 0
	}
	defer key.Close()
	return typ

}

//TODO: Доделать
func findParam(fullPath, value string, param []string) bool {
	typ := getValueType(fullPath, value)
	if typ == 0 {
		fmt.Println("param is not exist")
		return false
	}

	return true
}

func replaceParams(fullPath, value string, param []string) string {
	argsToReplace := strings.Split(param[0], "=>")
	typ := getValueType(fullPath, value)
	if typ == 0 {
		fmt.Println("param is not exist")
		return ""
	}

	srcInterface, _ := getParams(fullPath, value)
	switch typ {
	case 1:
		src := srcInterface.(string)
		re := regexp.MustCompile(argsToReplace[0])
		ss := re.ReplaceAllString(src, argsToReplace[1])
		setParams(fullPath, value, []string{ss})
		return ss
	case 2:
		src := srcInterface.(string)
		re := regexp.MustCompile(argsToReplace[0])
		ss := re.ReplaceAllString(src, argsToReplace[1])
		setParams(fullPath, value, []string{ss})
		return ss
	case 4:
		fmt.Println("Wron type of value(you try set the )")
	case 11:
		fmt.Println("Wron type of value(you try set the )")
	case 7:
		fmt.Println("Wron type of value(you try set the )")
	}
	return ""

}

func setParams(fullPath, value string, param []string) error {
	typ := getValueType(fullPath, value)
	if typ == 0 {
		fmt.Println("param is not exist")
	}
	var err error
	key := openKey(fullPath, SET)
	switch typ {
	case 1:
		err = key.SetStringValue(value, param[0])
	case 2:
		err = key.SetExpandStringValue(value, param[0])
	case 4:
		i, err2 := strconv.ParseUint(param[0], 10, 32)
		if err2 != nil {
			return err2
		}
		err = key.SetDWordValue(value, uint32(i))
	case 11:
		i, err2 := strconv.ParseUint(param[0], 10, 64)
		if err2 != nil {
			return err2
		}
		err = key.SetQWordValue(value, uint64(i))
	case 7:
		err = key.SetStringsValue(value, param)
	}
	key.Close()
	return err

}

func getParams(fullPath, value string) (interface{}, uint32) {
	typ := getValueType(fullPath, value)
	if typ == 0 {
		return nil, 0
	}
	key := openKey(fullPath, GET)
	switch typ {
	case 1, 2:
		str, _, _ := key.GetStringValue(value)
		return str, typ
	case 4, 11:
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

func openKey(fullPath string, operation string) registry.Key {
	var k registry.Key
	var err error
	var access uint32
	switch operation {
	case "set":
		access = registry.QUERY_VALUE | registry.SET_VALUE
	case "get":
		access = registry.QUERY_VALUE
	}
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
		fmt.Println("Operation " + operation)
		log.Fatal(err)
	}

	return k

}
