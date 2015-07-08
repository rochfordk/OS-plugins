// Arguments:
/*
check_cinder_error_rate]# ./check_cinder_error_rate -H 87.44.1.140 -P 3306 -u root -p 'password' -h 120 -S error -w 5 -c 10 --extra-opts cinder_check.cfg

-H hostname
-P port
-u user
-p password
-h hours
-S state 
-w warning
-c critical
--extra-opts

Extra Opts file used to pass in additional status checks and to declare endpoints for metrics injection.
*/

// TODO

// Help message and defaults for arguments
// proper error handling
// modify to handle multiple states
// add metrics output

package main

import (
          //"os"
          //"math"
    "github.com/fractalcat/nagiosplugin"
    "database/sql"
    _"github.com/go-sql-driver/mysql"
    "fmt"
    "flag"
)

// Command line Parameters
var (
    hostname string
    port int
    user string
    password string
    hours int
    state string
    warning int
    critical int
    extra_opts string
)

func init() {
        flag.StringVar(&hostname, "H", "127.0.0.1", "Address of OpenStack (Cinder DB) MySQL host")
        flag.IntVar(&port, "P", 3306, "Port of OpenStack (Cinder DB) MySQL host")
        flag.StringVar(&user, "u", "monitoring_user", "User of OpenStack (Cinder DB) MySQL host")
        flag.StringVar(&password, "p", "", "Password of OpenStack (Cinder DB) MySQL host")
        flag.IntVar(&hours, "h", 120, "Password of OpenStack (Cinder DB) MySQL host")
        flag.StringVar(&state, "S", "error", "Volume state for which check is applied")
        flag.IntVar(&warning, "w", 5, "Percentage threshold defining warning state")
        flag.IntVar(&critical, "c", 10, "Percentage threshold defining critical state")
        flag.StringVar(&extra_opts, "extra-opts", "", "")
        flag.Parse()

        //var Usage = func() {
        //fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
        //PrintDefaults()
        //}
}

func main() {

    type volume_state_count struct {
        count int
        percentage float64 
    }

    var volume_states = make(map[string]volume_state_count)

    var (
        Volume_Status string
        Total int
        Percentage float64
    )   

    // Initialize the check - this will return an UNKNOWN result
    // until more results are added.
    check := nagiosplugin.NewCheck()
    // If we exit early or panic() we'll still output a result.
    defer check.Finish()

    // obtain data here
    connString := fmt.Sprint(user, ":", password, "@tcp(", hostname, ":", port, ")/cinder")
    db, err := sql.Open("mysql", connString)

    if err != nil {
        check.Exitf(nagiosplugin.UNKNOWN, fmt.Sprint("Could not create database connection: ", err.Error()))
    }
    defer db.Close()

    // Open doesn't open a connection. Validate DSN data:
    err = db.Ping()
    if err != nil {
         check.Exitf(nagiosplugin.UNKNOWN, fmt.Sprint("Could not open database connection: ", err.Error()))
    }
    
    // Prepare statement for reading data
    stmt, err := db.Prepare("SELECT status AS `Volume_Status`, COUNT(1) AS `Total` ,COUNT(1) / t.cnt * 100 AS `Percentage` FROM volumes v CROSS JOIN (SELECT COUNT(1) AS cnt FROM volumes WHERE created_at > DATE_SUB(NOW(), INTERVAL ? HOUR)) t WHERE v.created_at > DATE_SUB(NOW(), INTERVAL ? HOUR) GROUP BY v.status;")
    if err != nil {
        check.Exitf(nagiosplugin.UNKNOWN, fmt.Sprint("Could not prepare statement: ", err.Error()))
    }
    defer stmt.Close()

    // Query the results for the last n hours
    rows, err := stmt.Query(hours, hours)
    if err != nil {
        check.Exitf(nagiosplugin.UNKNOWN, fmt.Sprint("Could not execute query: ", err.Error()))
    }
    defer rows.Close()

    for rows.Next(){
        err = rows.Scan(&Volume_Status, &Total, &Percentage)
        if err != nil {
                check.Exitf(nagiosplugin.UNKNOWN, fmt.Sprint("Invalid result set: ", err.Error()))
        }
        volume_states[Volume_Status] = volume_state_count {count: Total, percentage: Percentage} 
    }

    /*fmt.Println("Database Results:")
    for key, value := range volume_states {
        fmt.Println("Key:", key, "Value:", value)
    }*/

    //check for state
    if state_count, ok := volume_states[state]; ok {
        if state_count.percentage < float64(warning) {
            check.AddResult(nagiosplugin.OK, "Cinder Volume OK")
            check.AddPerfDatum(fmt.Sprint("Volumes in state '",state,"'"), "%", state_count.percentage, float64(warning), float64(critical), 0.0, 100.0)
            check.AddPerfDatum("Count", "", float64(state_count.count), 0.0 , 0.0 , 0.0 , 0.0 )
            check.Finish()
        } else if state_count.percentage >= float64(critical) {
            check.AddResult(nagiosplugin.CRITICAL, "Cinder Volume CRITICAL")
            check.AddPerfDatum(fmt.Sprint("Volumes in state '",state,"'"), "%", state_count.percentage, float64(warning), float64(critical), 0.0, 100.0)
            check.AddPerfDatum("Count", "", float64(state_count.count), 0.0 , 0.0 , 0.0 , 0.0 )
            check.Finish()
        } else {
            check.AddResult(nagiosplugin.WARNING, "Cinder Volume WARNING")
            check.AddPerfDatum(fmt.Sprint("Volumes in state '",state,"'"), "%", state_count.percentage, float64(warning), float64(critical), 0.0, 100.0)
            check.AddPerfDatum("Count", "", float64(state_count.count), 0.0 , 0.0 , 0.0 , 0.0 )
            check.Finish()
        }

    }else { // if the map doesn't contain the state key then no volumes are in that state and therefore non exceed the threshold
        check.AddResult(nagiosplugin.OK, "Cinder Volume OK")
        check.AddPerfDatum(fmt.Sprint("Volumes in state '",state,"'"), "%", state_count.percentage, float64(warning), float64(critical), 0.0, 100.0)
        check.AddPerfDatum("Count", "", float64(state_count.count), 0.0 , 0.0 , 0.0 , 0.0 )
        check.Finish()
    }

}

