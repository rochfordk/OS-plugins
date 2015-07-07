package main

import (
          "os"
          "math"
          "github.com/fractalcat/nagiosplugin"
	      "database/sql"
          _ "github.com/go-sql-driver/mysql"
          //"fmt"
)

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

    //fmt.Println("The square number of 13 is: %d", "120")

    // Initialize the check - this will return an UNKNOWN result
    // until more results are added.
    check := nagiosplugin.NewCheck()
    // If we exit early or panic() we'll still output a result.
    defer check.Finish()

    // obtain data here
    db, err := sql.Open("mysql", os.Args[1])
    if err != nil {
        panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
    }
    defer db.Close()

    // Open doesn't open a connection. Validate DSN data:
    err = db.Ping()
    if err != nil {
        panic(err.Error()) // proper error handling instead of panic in your app
    }

    
    // Use the DB normally, execute the querys etc
    // Prepare statement for reading data
    stmt, err := db.Prepare("SELECT status AS `Volume_Status`, COUNT(1) AS `Total` ,COUNT(1) / t.cnt * 100 AS `Percentage` FROM volumes v CROSS JOIN (SELECT COUNT(1) AS cnt FROM volumes WHERE created_at > DATE_SUB(NOW(), INTERVAL ? HOUR)) t WHERE v.created_at > DATE_SUB(NOW(), INTERVAL ? HOUR) GROUP BY v.status;")
    if err != nil {
        panic(err.Error()) // proper error handling instead of panic in your app
    }
    defer stmt.Close()

    //var rows Row
    // Query the results for the last n hours
    rows, err := stmt.Query(120, 120)
    if err != nil {
        panic(err.Error()) // proper error handling instead of panic in your app
    }
    defer rows.Close()

    for rows.Next(){
        err = rows.Scan(&Volume_Status, &Total, &Percentage)
        if err != nil {
                //t.Fatalf("Scan: %v", err)
                panic(err.Error()) // proper error handling instead of panic in your app
        }
        //vsc := volume_state_count {count: Total, percentage: Percentage} 
        //volume_states[Volume_Status] = vsc
        volume_states[Volume_Status] = volume_state_count {count: Total, percentage: Percentage} 
    }

    
    

    // Add an 'OK' result - if no 'worse' check results have been
    // added, this is the one that will be output.
    check.AddResult(nagiosplugin.OK, "everything looks shiny, cap'n")
    // Add some perfdata too (label, unit, value, min, max,
    // warn, crit). The math.Inf(1) will be parsed as 'no
    // maximum'.
    check.AddPerfDatum("badness", "kb", 3.14159, 0.0, math.Inf(1), 8000.0, 9000.0)

    // Parse an range from the command line and the more severe
    // results if they match.
    warnRange, err := nagiosplugin.ParseRange( "1:2" )
    if err != nil {
        check.AddResult(nagiosplugin.UNKNOWN, "error parsing warning range")
    }
    if warnRange.Check( 3.14159 ) {
        check.AddResult(nagiosplugin.WARNING, "Are we crashing again?")
    }
}

// put the data from the SQL query into a map of structs
// perfrom the logic operatinos and checking on that data structure

//http://stackoverflow.com/questions/17265463/how-do-i-convert-a-database-row-into-a-struct-in-go
// 