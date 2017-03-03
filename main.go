package main

import (
	"database/sql"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	DEBUG    = 1
	CONFIG   = "./greets.cfg"
	CALLTMPL = "./call.tmpl"
	CALLFILE = "./call_"
	LOGFILE  = "greetcall.log"
)

type CfgVars struct {
	LogFile   string
	CallFile  string
	CallTmpl  string
	MysqlHost string
	MysqlUser string
	MysqlPass string
	MysqlBase string
}

type Phone struct {
	PHONE string
}

var configfile string
var cfgvars CfgVars
var db *sql.DB

func init() {
	var cfgRaw = make(map[string]string)
	flag.StringVar(&configfile, "config", CONFIG, "Read configuration from this file")
	flag.StringVar(&configfile, "c", CONFIG, "Read configuration from this file (short)")
	flag.Parse()

	rawBytes, err := ioutil.ReadFile(configfile)
	if err != nil {
		log.Fatal(err)
	}

	text := string(rawBytes)
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		fields := strings.Split(line, "=")
		if len(fields) == 2 && strings.HasPrefix(fields[0], ";") == false {
			cfgRaw[strings.TrimSpace(fields[0])] = strings.TrimSpace(fields[1])
		}
	}

	if DEBUG == 1 {
		log.Println(cfgRaw, len(cfgRaw))
	}

	if len(cfgRaw) > 0 {
		if cfgRaw["logfile"] != "" {
			cfgvars.LogFile = cfgRaw["logfile"]
		} else {
			cfgvars.LogFile = LOGFILE
		}

		if cfgRaw["callfile"] != "" {
			cfgvars.CallFile = cfgRaw["callfile"]
		} else {
			cfgvars.CallFile = CALLFILE
		}

		if cfgRaw["calltmpl"] != "" {
			cfgvars.CallTmpl = cfgRaw["calltmpl"]
		} else {
			cfgvars.CallTmpl = CALLTMPL
		}

		if cfgRaw["mysql_host"] != "" {
			cfgvars.MysqlHost = cfgRaw["mysql_host"]
		} else {
			cfgvars.MysqlHost = "127.0.0.1"
		}

		if cfgRaw["mysql_user"] != "" {
			cfgvars.MysqlUser = cfgRaw["mysql_user"]
		} else {
			cfgvars.MysqlUser = ""
		}

		if cfgRaw["mysql_pass"] != "" {
			cfgvars.MysqlPass = cfgRaw["mysql_pass"]
		} else {
			cfgvars.MysqlPass = ""
		}

		if cfgRaw["mysql_base"] != "" {
			cfgvars.MysqlBase = cfgRaw["mysql_base"]
		} else {
			cfgvars.MysqlBase = "db"
		}
	}

}

func main() {
	/* связываем вывод log-сообщений с файлом */
	logTo := os.Stderr
	var err error
	if logTo, err = os.OpenFile(cfgvars.LogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600); err != nil {
		log.Fatal(err)
	}
	defer logTo.Close()
	log.SetOutput(logTo)

	/* устанавливаем соединение с БД */
	db, err = sql.Open("mysql", cfgvars.MysqlUser+":"+cfgvars.MysqlPass+"@tcp("+cfgvars.MysqlHost+":3306)/"+cfgvars.MysqlBase)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	if DEBUG == 1 {
		log.Println("Connect to SQL " + cfgvars.MysqlHost + "/" + cfgvars.MysqlBase + " success")
	}

	/* основной цикл обработки соединений */
	//for {
	generateCalls()
	time.Sleep(500 * time.Millisecond)
	//}
}

/* Основной обработчик */
func generateCalls() {

	sql := "select * from testcalls where status='new' group by pnumber"
	log.Println(sql)
	rows, err := db.Query(sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var pnumber string
		err = rows.Scan(&id, &pnumber, nil, nil, nil)
		if err != nil {
			log.Println(err)
		}

		num := Phone{pnumber}
		fname := cfgvars.CallFile + pnumber
		callF, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE, 0600)

		if err != nil {
			log.Fatal(err)
		}
		defer callF.Close()

		callT, err := template.ParseFiles(cfgvars.CallTmpl)
		if err != nil {
			log.Println(err)
		}

		err = callT.Execute(callF, num)
		callF.Chown(109, 114)
		if err != nil {
			log.Println(err)
		}
	}

}
