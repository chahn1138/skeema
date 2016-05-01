package main

import (
	"fmt"
	//"io/ioutil"
	"os"
	"path"
	//"path/filepath"
	//"strings"

	"github.com/skeema/tengo"
	"github.com/spf13/pflag"
)

func init() {
	initFlags := pflag.NewFlagSet("init", pflag.ExitOnError)
	initFlags.String("alias", "<host>", "Override the directory name to use for a host, or supply explicit blank string to put schema subdirs at the top level")

	long := `Creates a filesystem representation of the schemas and tables on a db host. For
each schema on the host (or just the single schema specified by --schema), a
directory with a .skeema config file will be created. Each directory will be
populated with .sql files containing CREATE TABLE statements for every table
in the schema.`
	Commands["init"] = Command{
		Name:    "init",
		Short:   "Save a live db's schemas and tables to the filesystem",
		Long:    long,
		Flags:   initFlags,
		Handler: InitCommand,
	}
}

func InitCommand(input *pflag.FlagSet, cliConfig ParsedGlobalFlags) {
	// Figure out what dir path we're using
	dirPath, err := cliConfig.DirPath(true)
	if err != nil {
		fmt.Println("Invalid --dir option:", err)
		os.Exit(1)
	}

	cfg := NewConfig(dirPath)
	target := cfg.TargetList("master", &cliConfig)[0] // TODO: handle branches appropriately, and ditto for multiple targets

	alias, _ := input.GetString("alias")
	var instancePath string
	if alias == "<host>" {
		instancePath = path.Join(dirPath, target.HostAndOptionalPort())
	} else if alias == "." {
		instancePath = dirPath
	} else {
		instancePath = path.Join(dirPath, alias)
	}

	// YOUAREHERE: create a dir for the instance, unless it already exists
	fi, err := os.Stat(instancePath)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println("Unable to use specified directory: ", err)
			os.Exit(1)
		}
		err = os.MkdirAll(instancePath, 0777)
		if err != nil {
			fmt.Println("Unable to create specified directory: ", err)
			os.Exit(1)
		}
	} else if !fi.IsDir() {
		fmt.Printf("Path %s already exists but is not a directory\n", instancePath)
		os.Exit(1)
	} else {
		// dir already exists -- for now let's bail
		// TODO: implement updating of existing dir?!
		fmt.Printf("Path %s already exists, refusing to overwrite!\n", instancePath)
		os.Exit(1)
	}

	driver := "mysql"
	instance := &tengo.Instance{Driver: driver, DSN: target.DSN()}

	// iterate over the schemas; create a dir w/ .skeema and .sql files
	var schemas []*tengo.Schema
	if target.Schema != "" {
		if !instance.HasSchema(target.Schema) {
			fmt.Printf("Schema %s does not exist in this instance\n", target.Schema)
			os.Exit(1)
		}
		schemas = []*tengo.Schema{instance.Schema(target.Schema)}
	} else {
		schemas = instance.Schemas()
	}

	for _, s := range schemas {
		fmt.Println("found schema: ", s.Name) // TODO
	}
}