package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hidu/mysql-schema-sync/internal"
)

var configPath = flag.String("conf", "./config.json", "json config file path")
var sync = flag.Bool("sync", false, "sync schema changes to dest's db\non default, only show difference")
var drop = flag.Bool("drop", false, "drop fields,index,foreign key only on dest's table")

var source = flag.String("source", "", "sync from, eg: test@(10.10.0.1:3306)/my_online_db_name\nwhen it is not empty,[-conf] while ignore")
var dest = flag.String("dest", "", "sync to, eg: test@(127.0.0.1:3306)/my_local_db_name")
var tables = flag.String("tables", "", "tables to sync\neg : product_base,order_*")
var tablesIgnore = flag.String("tables_ignore", "", "tables ignore sync\neg : product_base,order_*")
var mailTo = flag.String("mail_to", "", "overwrite config's email.to")
var singleSchemaChange = flag.Bool("single_schema_change", false, "single schema changes ddl command a single schema change")

func init() {
	log.SetFlags(log.Lshortfile | log.Ldate)
	df := flag.Usage
	flag.Usage = func() {
		df()
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "mysql schema sync tools "+internal.Version)
		fmt.Fprint(os.Stderr, internal.AppURL+"\n\n")
	}
}

var cfg *internal.Config

func main() {
	flag.Parse()
	if *source == "" {
		cfg = internal.LoadConfig(*configPath)
	} else {
		cfg = new(internal.Config)
		cfg.SourceDSN = *source
		cfg.DestDSN = *dest
	}
	cfg.Sync = *sync
	cfg.Drop = *drop
	cfg.SingleSchemaChange = *singleSchemaChange

	if *mailTo != "" && cfg.Email != nil {
		cfg.Email.To = *mailTo
	}

	if cfg.Tables == nil {
		cfg.Tables = []string{}
	}
	if cfg.TablesIgnore == nil {
		cfg.TablesIgnore = []string{}
	}
	if *tables != "" {
		_ts := strings.Split(*tables, ",")
		for _, _name := range _ts {
			_name = strings.TrimSpace(_name)
			if _name != "" {
				cfg.Tables = append(cfg.Tables, _name)
			}
		}
	}
	if *tablesIgnore != "" {
		_ts := strings.Split(*tablesIgnore, ",")
		for _, _name := range _ts {
			_name = strings.TrimSpace(_name)
			if _name != "" {
				cfg.TablesIgnore = append(cfg.TablesIgnore, _name)
			}
		}
	}
	defer (func() {
		if err := recover(); err != nil {
			log.Println(err)
			cfg.SendMailFail(fmt.Sprintf("%s", err))
			log.Fatalln("exit")
		}
	})()

	cfg.Check()
	internal.CheckSchemaDiff(cfg)
}
