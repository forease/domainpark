
package main

import (
    "fmt"
    //"io/ioutil"
    //"net"
    "net/http"
    "html/template"
    "flag"
    "strings"
    "os"
    "time"
    //"strings"
    redis "github.com/alphazero/Go-Redis"
    config "github.com/jonsen/goconfig/config"
)

const (
    cfgFileDefault = "./serverd.conf"
)

type sysCfg struct {
    // SiteName string
    logfile, reportTo string
    webPort, debug int

    // redis setup
    redisHost, redisAuth, redisKeyFix string
    redisPort, redisDb int

    // SMTP setup
    smtpHost, smtpUser, smtpPassword string
    smtpPort int
    smtpAuth, smtpTLS, smtpDaemon bool

}

var AppConfig sysCfg


// Read cfgfile or setup defaults.
func setupConfig() {
    // 初始化读取配置文件
    c, err := config.ReadDefault( *cfgfile )
    if err != nil {
        fmt.Println( "read config file error: ", *cfgfile, err )
        os.Exit(1)
    }

    // 读取[common]段配置

    AppConfig.webPort, _ = c.Int("common", "webport")
    AppConfig.debug, _ = c.Int("common", "debug")
    AppConfig.logfile, _ = c.String("common", "log")
    AppConfig.reportTo, _ = c.String("common", "reportto")

    AppConfig.redisHost, _ = c.String("redis", "host")
    AppConfig.redisAuth, _ = c.String("redis", "auth")
    AppConfig.redisKeyFix, _ = c.String("redis", "key_prefix")
    AppConfig.redisPort, _ = c.Int("redis", "prt")
    AppConfig.redisDb, _ = c.Int("redis", "db")

    // SMTP
    AppConfig.smtpHost, _ = c.String("smtp", "host")
    AppConfig.smtpUser, _ = c.String("smtp", "user")
    AppConfig.smtpPassword, _ = c.String("smtp", "password")
    AppConfig.smtpPort, _ = c.Int("smtp", "port")
    AppConfig.smtpAuth, _ = c.Bool("smtp", "auth")
    AppConfig.smtpTLS, _ = c.Bool("smtp", "tls")
    AppConfig.smtpDaemon, _ = c.Bool("smtp", "daemon" )

}

func redisConnect() ( rd redis.Client, err error ) {
    spec := redis.DefaultSpec().Db(AppConfig.redisDb)
    if AppConfig.redisHost != "" {
        spec.Host( AppConfig.redisHost )
    }
    if AppConfig.redisAuth != "" {
        spec.Password( AppConfig.redisAuth )
    }
    port := AppConfig.redisPort
    if port > 0 && port < 65535 {
        spec.Port( port )
    }

    rd, err = redis.NewSynchClientWithSpec(spec)
    if err != nil {
        fmt.Printf( "Connect Redis: %s", err )
        return rd, err
    }

    return rd, nil
}


func rediClose() {

}

func yesterday() string {

    y := time.Now().Unix() - 86400

    n := time.Unix( y, 0 ).Format("20060102")

    //fmt.Println(y, n)

    return n
}

func report() (rep string, err error) {
    rd, err := redisConnect()
    if err != nil {
        fmt.Println(err)
        return
    }

    defer rd.Quit()

    date := yesterday() //time.Now().Format("20060102")
    keys := "*:" + date

    rep = "HI\n"
    rep += "    This report for " + date + "\n"
    rep += "    Domain and count:\n\n"

    list, err := rd.Keys( keys )
    if err != nil {
        return "", err
    }
    // fmt.Println(list)    

    for _, v := range list {
        count, _ := rd.Get(v)
        rep += "    " + v + ": " + string(count) + "\n";
    }

    return rep, nil


}

func reportServer() {
    var day int

    fmt.Println("running report server ...")

    for {
        currDay := time.Now().Day()
        //fmt.Println( currDay, day )
        if day != currDay {
            // execute send mail ...
            fmt.Println("send mail now ...")
            reportBody, err := report()
            if err == nil {
                err = MailSender( "Domain Park Report " +
                time.Now().Format("2006-01-02"), reportBody, AppConfig.reportTo )
                if err != nil {
                    fmt.Println("send mail err", err )
                } else {
                    //fmt.Println("send mail ok")
                }
            }

            day = currDay
        }
        time.Sleep( 30 * time.Minute )
    }

}

/**
 * 连接入口
 *
 * author: jonsen yang
 * date: 2012-10-17
 */
func makeHandler(w http.ResponseWriter, r *http.Request) {
    if r.RequestURI == "/favicon.ico" || r.RequestURI == "/robots.txt" {
        //http.NotFound(w,r)
        return
    }
    fmt.Printf( "[%v] %s %s %s %s %v\n",time.Now(), r.RemoteAddr, r.Method, r.Host, r.RequestURI, r.Referer()  )

    t, _ := template.ParseFiles("index.html")
    t.Execute(w, nil)



    hosts := strings.Split( r.Host, ":" )
//    fmt.Println(r.Host, hosts, r.RequestURI )
//    fmt.Println( getDomainRoot( hosts[0] ) )


    domain, rootid := getDomainRoot( strings.ToLower( hosts[0] ))
    if domain == "" || rootid < 1 {
        //fmt.Println( domain, rootid )
        return
    }

    rd, err := redisConnect()
    if err != nil {
        fmt.Println(err)
        return
    }

    date := time.Now().Format("20060102")
    fmt.Println( domain + ":" + date)

    key := domain + ":" + date
    rd.Incr( key )

    rd.Quit()

}



var cfgfile = flag.String("c", cfgFileDefault, "Config file")


func main() {
    flag.Parse()
    setupConfig()
    initDomainExt()

    yesterday()

    go reportServer()

    listen := fmt.Sprintf(":%d", AppConfig.webPort )

    http.HandleFunc( "/", makeHandler )
    err := http.ListenAndServe( listen, nil )
    if err != nil {
        fmt.Println( "Run server err", err )
        os.Exit(1)
    }

}


