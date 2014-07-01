package main

import (
  "fmt"
  "io/ioutil"
  "net/http"
  "net/url"
  //"html/template"
  "github.com/PuerkitoBio/goquery"
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
 <div>
 <form action="https://api.weibo.com/oauth2/authorize">
  <input type="hidden" name="client_id" value="2909741328">
  <input type="hidden" name="redirect_uri" value="http://ambient-depth-612.appspot.com/connect-weibo">
  <input type="hidden" name="response_type" value="code">
  <button>View</button>
  </form></div>
 </body>
 </html>
`

func connect_weibo(w http.ResponseWriter, r *http.Request) {
  resp, err := http.PostForm(
    "https://api.weibo.com/oauth2/access_token",
    url.Values{"client_id":{"2909741328"},
    "client_secret":{"61d75f2cceb35c8d95dd7185afb1dd5c"},
    "grant_type":{"authorization_code"},
    "code":{r.FormValue("code")},
    "redirect_uri":{"http://ambient-depth-612.appspot.com/view"}})

  body, err := ioutil.ReadAll(resp.Body)
  if (resp.StatusCode != http.StatusOK) {
    fmt.Fprint(w, "fatal1\n")
    return
  }

  if err !=nil {
    fmt.Fprint(w, "fatal2\n")
    return
  }

  if (body[0] == 0) {
    fmt.Fprint(w, "connected1\n")
  } else {
    fmt.Fprint(w, "connected2\n")
  }

}

func view(w http.ResponseWriter, r *http.Request) {
  fmt.Fprint(w, "view")

  // qq by scraping
  doc, _ := goquery.NewDocument("e.t.qq.com")
  fmt.Fprint(w, doc.Find("hoge"))
}
