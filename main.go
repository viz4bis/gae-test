package main

import (
  "fmt"
  "net/http"
  "net/url"
  //"html/template"
  )

func init() {
  http.HandleFunc("/", root)
  http.HandleFunc("/connect-weibo", connect_weibo)
  http.HandleFunc("/view", view)
}

func root(w http.ResponseWriter, r *http.Request) {
  fmt.Fprint(w, root_html)
}

const root_html = `
 <html> 
 <body>
 <div><form action="https://api.weibo.com/oauth2/authorize?client_id=2909741328&redirect_uri=http://ambient-depth-612.appspot.com/connect-weibo&response_type=code"><button>View</button></form></div>
 </body>
 </html>
`

func connect(w http.ResponseWriter, r *http.Request) {
  http.PostForm("https://api.weibo.com/oauth2/access_token",
    url.Values{"client_id":{"2909741328"},
    "client_secret":{"61d75f2cceb35c8d95dd7185afb1dd5c"},
    "grant_type":{"authorization_code"},
    "code":{r.FormValue("code")},
    "redirect_uri":{"http://ambient-depth-612.appspot.com/connected"}})
}

func connected(w http.ResponseWriter, r *http.Request) {
  fmt.Fprint(w, "connected")
}
