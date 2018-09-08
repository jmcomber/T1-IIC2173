// credits to: https://medium.com/@martinomburajr/building-a-go-web-app-from-scratch-to-deploying-on-google-cloud-part-3a-deploying-our-go-app-16f5a2f44634


package main
import (
	"net/http"
	"fmt"
	"time"
	"html/template"
	"database/sql"
	"os"

  _ "github.com/lib/pq"

)

//Create a struct that holds information to be displayed in our HTML file
type Welcome struct {
	Comment string
	Time string
	IP string
}



var (
	  host     = "localhost"
	  port     = 5432
	  user     = "postgres"
	  password = ""
	  dbname   = ""
	)



//Go application entrypoint
func main() {
	password = os.Args[1]
	dbname = os.Args[2]
	fmt.Println(password);
	fmt.Println(dbname);

	
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
    "password=%s dbname=%s sslmode=disable",
    host, port, user, password, dbname)
    db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
	  	msg := fmt.Sprintf("Could not open database: %v", err)
	  	http.HandleFunc("/" , func(w http.ResponseWriter, r *http.Request) {
	  		http.Error(w, msg, http.StatusInternalServerError)
	    	return
	  	})
	}
	defer db.Close()






	//Instantiate a Welcome struct object and pass in some random information.
	//We shall get the name of the user as a query parameter from the URL
	welcome := Welcome{"Anonymous", time.Now().Format(time.Stamp), "-"}

	//We tell Go exactly where we can find our html file. We ask Go to parse the html file (Notice
	// the relative path). We wrap it in a call to template.Must() which handles any errors and halts if there are fatal errors

	var templates = template.Must(template.ParseGlob("templates/*"))


	//Our HTML comes with CSS that go needs to provide when we run the app. Here we tell go to create
	// a handle that looks in the static directory, go then uses the "/static/" as a url that our
	//html can refer to when looking for our css and other files.

	http.Handle("/static/", //final url can be anything
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static")))) //Go looks in the relative static directory first, then matches it to a
			//url of our choice as shown in http.Handle("/static/"). This url is what we need when referencing our css files
			//once the server begins. Our html code would therefore be <link rel="stylesheet"  href="/static/stylesheet/...">
			//It is important to note the final url can be whatever we like, so long as we are consistent.

	//This method takes in the URL path "/" and a function that takes in a response writer, and a http request.
	http.HandleFunc("/" , func(w http.ResponseWriter, r *http.Request) {

		//Takes the name from the URL query e.g ?name=Martin, will set welcome.Name = Martin.
		if comment := r.FormValue("comment"); comment != "" {
			welcome.Comment = comment;
		}
		welcome.IP = r.RemoteAddr;


		if (welcome.Comment != "" && welcome.Comment != "Anonymous") {
			welcome.Time = time.Now().Format(time.Stamp)
			welcome.IP = r.RemoteAddr;
			stmt := "INSERT INTO comments (IP, time, comment) VALUES ($1, $2, $3)"
	        _, err := db.Exec(stmt, welcome.Comment, welcome.Time, welcome.IP)
	        welcome.Comment = "";

			if err != nil {
	            msg := fmt.Sprintf("Could not save comment: %v", err)
	            http.Error(w, msg, http.StatusInternalServerError)
	            return
		    }

		}
		
	    rows, err := db.Query("SELECT * FROM comments ORDER BY time DESC")
        if err != nil {
            // return nil, fmt.Errorf("Could not get recent comments: %v", err)
            fmt.Println("Could not get recent comments: %v", err)
        }
        defer rows.Close()

        var welcomes []Welcome
        for rows.Next() {
            var v Welcome
            if err := rows.Scan(&v.Comment, &v.Time, &v.IP); err != nil {
                // return nil, fmt.Errorf("Could not get values out of row: %v", err)
                fmt.Println("Could not get values out of row: %v", err)
            }
            welcomes = append(welcomes, v)
        }


		// welcomes, err := queryComments()
		if err != nil {
			msg := fmt.Sprintf("Could not show comments: %v", err)
            http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		

		//If errors show an internal server error message
		//I also pass the welcome struct to the welcome-template.html file.
		if err := templates.ExecuteTemplate(w, "home", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if err := templates.ExecuteTemplate(w, "comments", welcomes); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}



	})

	//Start the web server, set the port to listen to 8080. Without a path it assumes localhost
	//Print any errors from starting the webserver using fmt
	fmt.Println("Listening on Port 80");
	// ln, err := net.Listen("tcp", ":80");
	fmt.Println(http.ListenAndServe(":8080", nil));
}
